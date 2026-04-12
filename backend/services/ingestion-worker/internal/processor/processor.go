package processor

import (
	"context"
	"log/slog"
	"time"

	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/kafka"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/repo"
)

// Processor handles document ingestion processing
type Processor struct {
	docRepo           *repo.DocumentsRepository
	processingSleepmMS int
	logger            *slog.Logger
}

// NewProcessor creates a new processor
func NewProcessor(docRepo *repo.DocumentsRepository, processingSleepmMS int, logger *slog.Logger) *Processor {
	return &Processor{
		docRepo:           docRepo,
		processingSleepmMS: processingSleepmMS,
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
	)

	// Simulate processing delay (skeleton - replace with actual ingestion logic)
	if p.processingSleepmMS > 0 {
		select {
		case <-time.After(time.Duration(p.processingSleepmMS) * time.Millisecond):
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
