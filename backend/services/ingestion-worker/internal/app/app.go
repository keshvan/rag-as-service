package app

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/config"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/kafka"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/processor"
)

type App struct {
	cfg      *config.Config
	logger   *slog.Logger
	consumer *kafka.Consumer
	proc     *processor.Processor
}

func NewApp(cfg *config.Config, logger *slog.Logger, consumer *kafka.Consumer, proc *processor.Processor) *App {
	if cfg.WorkerConcurrency <= 0 {
		cfg.WorkerConcurrency = 4
	}

	return &App{
		cfg:      cfg,
		logger:   logger,
		consumer: consumer,
		proc:     proc,
	}
}

func (a *App) Run() error {
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
