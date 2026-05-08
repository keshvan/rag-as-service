package main

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
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

func main() {
	// Setup logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Load configuration
	cfg := config.LoadConfig()

	// Validate required config
	if cfg.DatabaseURL == "" {
		logger.Error("DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	// Connect to database
	logger.Info("Connecting to database", "database_url", maskDatabaseURL(cfg.DatabaseURL))
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to create connection pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		logger.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}
	logger.Info("Connected to database")

	// Initialize repositories
	docRepo := repo.NewDocumentsRepository(pool)

	downloader, err := storage.NewS3Downloader(
		context.Background(),
		cfg.S3Endpoint,
		cfg.S3Region,
		cfg.S3Bucket,
		cfg.S3AccessKeyID,
		cfg.S3SecretAccessKey,
		cfg.DownloadDir,
	)
	if err != nil {
		logger.Error("Failed to create S3 downloader", "error", err)
		os.Exit(1)
	}

	// Initialize embedder from pkg/common/embeddings
	embedder, err := embeddings.NewYandexAIClient(embeddings.YandexAIConfig{
		ApiKey:   cfg.YandexApiKey,
		FolderID: cfg.YandexFolderID,
		BaseURL:  cfg.YandexBaseURL,
	})
	if err != nil {
		logger.Error("Failed to create Yandex AI embedder", "error", err)
		os.Exit(1)
	}

	// Initialize Qdrant vector store
	qdrant, err := qdrant.NewQdrant(qdrant.QdrantConfig{
		Host:       cfg.QdrantHost,
		Port:       cfg.QdrantPort,
		Collection: cfg.QdrantCollection,
		UseTLS:     cfg.QdrantUseTLS,
	})
	if err != nil {
		logger.Error("Failed to create Qdrant client", "error", err)
		os.Exit(1)
	}

	ext := extractor.NewExtractor()
	chunker := chunking.NewChunker(1000, 200)

	// Initialize processor
	proc := processor.NewProcessor(
		docRepo,
		downloader,
		ext,
		chunker,
		embedder,
		qdrant,
		logger,
	)

	// Create Kafka consumer
	logger.Info("Creating Kafka consumer",
		"brokers", cfg.KafkaBrokers,
		"topic", cfg.KafkaTopic,
		"group_id", cfg.KafkaGroupID,
	)
	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID, logger)
	defer consumer.Close()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Starting ingestion worker")

	// Message consumption loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-sigChan:
			logger.Info("Received shutdown signal, gracefully shutting down")
			cancel()
			return
		case <-ctx.Done():
			return
		default:
		}

		// Read message with timeout
		msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		msgWithEvent, err := consumer.ReadMessage(msgCtx)
		cancel()

		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				continue
			}
			logger.Error("Failed to read message", "error", err)
			continue
		}

		// Process the event
		procCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		if err := proc.Process(procCtx, msgWithEvent.Event); err != nil {
			logger.Error("Failed to process event",
				"document_id", msgWithEvent.Event.DocumentID,
				"organization_id", msgWithEvent.Event.OrganizationID,
				"error", err,
			)
			cancel()
			continue
		}
		cancel()

		// Commit offset on successful processing
		commitCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		if err := consumer.CommitMessages(commitCtx, msgWithEvent.Message); err != nil {
			logger.Error("Failed to commit message", "error", err)
		}
		cancel()
	}
}

// maskDatabaseURL masks the password in the database URL for logging
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
