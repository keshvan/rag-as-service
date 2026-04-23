package processor

import (
	"context"
	"log/slog"
	"time"

	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/kafka"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/repo"
)

// Processor handles document ingestion processing

type S3Downloader interface {
	DownloadToFile(ctx context.Context, objectKey, orgID, docID string) (string, error)
}

type Processor struct {
	docRepo           *repo.DocumentsRepository
	downloader        S3Downloader
	processingSleepMS int
	logger            *slog.Logger
}

// NewProcessor creates a new processor
func NewProcessor(docRepo *repo.DocumentsRepository, downloader S3Downloader, processingSleepMS int, logger *slog.Logger) *Processor {
	return &Processor{
		docRepo:           docRepo,
		downloader:        downloader,
		processingSleepMS: processingSleepMS,
		logger:            logger,
	}
}

// Process processes a document upload event
func (p *Processor) Process(ctx context.Context, event *kafka.DocumentUploadedEvent) error {
	// Check current status for idempotency
	status, err := p.docRepo.GetStatus(ctx, event.OrganizationID, event.DocumentID)
	if err != nil {
		p.logger.Error("Failed to get document status",
			"doc_id", event.DocumentID,
			"org_id", event.OrganizationID,
			"error", err,
		)
		return err
	}

	// If already indexed, skip processing (idempotency)
	if status == "indexed" {
		p.logger.Info("Document already indexed, skipping",
			"doc_id", event.DocumentID,
			"org_id", event.OrganizationID,
		)
		return nil
	}

	// Mark as processing
	if err := p.docRepo.MarkProcessing(ctx, event.OrganizationID, event.DocumentID); err != nil {
		p.logger.Error("Failed to mark document as processing",
			"doc_id", event.DocumentID,
			"org_id", event.OrganizationID,
			"error", err,
		)
		return err
	}

	p.logger.Info("Document marked as processing",
		"doc_id", event.DocumentID,
		"org_id", event.OrganizationID,
		"object_key", event.ObjectKey,
	)

	localPath, err := p.downloader.DownloadToFile(ctx, event.ObjectKey, event.OrganizationID, event.DocumentID)
	if err != nil {
		p.logger.Error("Failed to download document from S3",
			"doc_id", event.DocumentID,
			"org_id", event.OrganizationID,
			"object_key", event.ObjectKey,
			"error", err,
		)

		_ = p.docRepo.MarkFailed(ctx, event.OrganizationID, event.DocumentID, err.Error())
		return err
	}

	p.logger.Info("Document downloaded successfully",
		"doc_id", event.DocumentID,
		"org_id", event.OrganizationID,
		"local_path", localPath,
		"content_type", event.ContentType,
		"size_bytes", event.SizeBytes,
	)

	if p.processingSleepMS > 0 {
		select {
		case <-time.After(time.Duration(p.processingSleepMS) * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Mark as indexed
	if err := p.docRepo.MarkIndexed(ctx, event.OrganizationID, event.DocumentID); err != nil {
		p.logger.Error("Failed to mark document as indexed",
			"doc_id", event.DocumentID,
			"org_id", event.OrganizationID,
			"error", err,
		)
		if markFailErr := p.docRepo.MarkFailed(ctx, event.OrganizationID, event.DocumentID, err.Error()); markFailErr != nil {
			p.logger.Error("Failed to mark document as failed",
				"doc_id", event.DocumentID,
				"org_id", event.OrganizationID,
				"error", markFailErr,
			)
		}
		return err
	}

	p.logger.Info("Document indexed successfully",
		"doc_id", event.DocumentID,
		"org_id", event.OrganizationID,
	)

	return nil
}
