package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keshvan/rag-as-service/backend/pkg/common/embeddings"
	"github.com/keshvan/rag-as-service/backend/pkg/common/qdrant"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/chunking"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/config"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/extractor"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/kafka"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/processor"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/repo"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/storage"
)

type App struct {
	cfg      *config.Config
	logger   *slog.Logger
	pool     *pgxpool.Pool
	consumer *kafka.Consumer
	proc     *processor.Processor
}

func NewApp(cfg *config.Config) (*App, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if cfg.WorkerConcurrency <= 0 {
		cfg.WorkerConcurrency = 4
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL environment variable is required")
	}

	ctx := context.Background()

	logger.Info("Connecting to database", "database_url", maskDatabaseURL(cfg.DatabaseURL))
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	logger.Info("Connected to database")

	docRepo := repo.NewDocumentsRepository(pool)

	downloader, err := storage.NewS3Downloader(
		ctx,
		cfg.S3Endpoint,
		cfg.S3Region,
		cfg.S3Bucket,
		cfg.S3AccessKeyID,
		cfg.S3SecretAccessKey,
		cfg.DownloadDir,
	)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("create S3 downloader: %w", err)
	}

	embedder, err := embeddings.NewYandexAIEmbedder(embeddings.YandexAIConfig{
		ApiKey:   cfg.YandexApiKey,
		FolderID: cfg.YandexFolderID,
		BaseURL:  cfg.YandexBaseURL,
	})
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("create Yandex AI embedder: %w", err)
	}

	vectorstore, err := qdrant.NewQdrant(qdrant.QdrantConfig{
		Host:       cfg.QdrantHost,
		Port:       cfg.QdrantPort,
		Collection: cfg.QdrantCollection,
		UseTLS:     cfg.QdrantUseTLS,
	})
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("create Qdrant client: %w", err)
	}

	ext := extractor.NewExtractor()
	chunker := chunking.NewChunker(1000, 200)

	proc := processor.NewProcessor(
		docRepo,
		downloader,
		ext,
		chunker,
		embedder,
		vectorstore,
		logger,
	)

	logger.Info("Creating Kafka consumer",
		"brokers", cfg.KafkaBrokers,
		"topic", cfg.KafkaTopic,
		"group_id", cfg.KafkaGroupID,
	)
	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID, logger)

	return &App{
		cfg:      cfg,
		logger:   logger,
		pool:     pool,
		consumer: consumer,
		proc:     proc,
	}, nil
}

func (a *App) Run() error {
	defer a.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	jobs := make(chan *kafka.MessageWithEvent, a.cfg.WorkerConcurrency*2)

	var wg sync.WaitGroup
	for i := 0; i < a.cfg.WorkerConcurrency; i++ {
		wg.Add(1)
		go a.runWorker(ctx, &wg, i+1, jobs)
	}

	a.logger.Info("Starting ingestion worker",
		"concurrency", a.cfg.WorkerConcurrency,
		"topic", a.cfg.KafkaTopic,
		"group_id", a.cfg.KafkaGroupID,
	)

readLoop:
	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Shutdown signal received, stopping message intake")
			break readLoop
		default:
		}

		msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		msgWithEvent, err := a.consumer.ReadMessage(msgCtx)
		cancel()

		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				if ctx.Err() != nil {
					break readLoop
				}
				continue
			}

			a.logger.Error("Failed to read message", "error", err)
			continue
		}

		select {
		case jobs <- msgWithEvent:
		case <-ctx.Done():
			break readLoop
		}
	}

	close(jobs)
	wg.Wait()

	a.logger.Info("Ingestion worker stopped")
	return nil
}

func (a *App) Close() {
	if a.consumer != nil {
		if err := a.consumer.Close(); err != nil {
			a.logger.Error("Failed to close Kafka consumer", "error", err)
		}
	}

	if a.pool != nil {
		a.pool.Close()
	}
}

func (a *App) runWorker(ctx context.Context, wg *sync.WaitGroup, workerID int, jobs <-chan *kafka.MessageWithEvent) {
	defer wg.Done()

	for msgWithEvent := range jobs {
		procCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		err := a.proc.Process(procCtx, msgWithEvent.Event)
		cancel()

		if err != nil {
			a.logger.Error("Failed to process event",
				"worker_id", workerID,
				"document_id", msgWithEvent.Event.DocumentID,
				"organization_id", msgWithEvent.Event.OrganizationID,
				"error", err,
			)
			continue
		}

		commitCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = a.consumer.CommitMessages(commitCtx, msgWithEvent.Message)
		cancel()

		if err != nil {
			a.logger.Error("Failed to commit message",
				"worker_id", workerID,
				"document_id", msgWithEvent.Event.DocumentID,
				"organization_id", msgWithEvent.Event.OrganizationID,
				"error", err,
			)
			continue
		}

		a.logger.Info("Message committed",
			"worker_id", workerID,
			"document_id", msgWithEvent.Event.DocumentID,
			"organization_id", msgWithEvent.Event.OrganizationID,
		)
	}
}

func maskDatabaseURL(dbURL string) string {
	u, err := url.Parse(dbURL)
	if err != nil {
		return "***"
	}
	if u.User != nil {
		u.User = url.User(u.User.Username())
	}
	return u.String()
}
