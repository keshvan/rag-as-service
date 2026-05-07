package processor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/keshvan/rag-as-service/backend/pkg/common/qdrant"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/chunking"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/extractor"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/kafka"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/repo"
)

// Processor handles document ingestion processing

type S3Downloader interface {
	DownloadToFile(ctx context.Context, objectKey, orgID, docID string) (string, error)
}

type Embedder interface {
	Embed(ctx context.Context, texts []string, textType string) ([][]float32, error)
}

type Processor struct {
	docRepo           *repo.DocumentsRepository
	downloader        S3Downloader
	extractor         *extractor.Extractor
	chunker           *chunking.Chunker
	embedder          Embedder
	vectorstore       *qdrant.Qdrant
	processingSleepMS int
	logger            *slog.Logger
}

// NewProcessor creates a new processor
func NewProcessor(
	docRepo *repo.DocumentsRepository,
	downloader S3Downloader,
	extractor *extractor.Extractor,
	chunker *chunking.Chunker,
	embedder Embedder,
	vectorstore *qdrant.Qdrant,
	logger *slog.Logger,
) *Processor {
	return &Processor{
		docRepo:     docRepo,
		downloader:  downloader,
		extractor:   extractor,
		chunker:     chunker,
		embedder:    embedder,
		vectorstore: vectorstore,
		logger:      logger,
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

	//1. Extract
	text, err := p.extractor.Extract(ctx, localPath)
	if err != nil {
		p.logger.Error("Failed to extract text",
			"doc_id", event.DocumentID,
			"org_id", event.OrganizationID,
			"error", err,
		)
		_ = p.docRepo.MarkFailed(ctx, event.OrganizationID, event.DocumentID, "extract: "+err.Error())
		return fmt.Errorf("extract text: %w", err)
	}

	// 2. Chunk text
	chunks := p.chunker.Split(text)
	if len(chunks) == 0 {
		_ = p.docRepo.MarkFailed(ctx, event.OrganizationID, event.DocumentID, "no chunks")
		return fmt.Errorf("no chunks produced")
	}

	// 3. Collect chunk texts for batch embedding
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Text
	}

	// Log chunk order for debugging
	for i, chunk := range chunks {
		preview := chunk.Text
		if len([]rune(preview)) > 80 {
			preview = string([]rune(preview)[:80]) + "..."
		}
		p.logger.Info("Chunk order",
			"doc_id", event.DocumentID,
			"chunk_index", chunk.Index,
			"order", i,
			"len", len([]rune(chunk.Text)),
			"preview", preview,
		)
	}

	vectors, err := p.embedder.Embed(ctx, texts, "doc")
	if err != nil {
		p.logger.Error("Failed to generate embeddings",
			"doc_id", event.DocumentID,
			"org_id", event.OrganizationID,
			"error", err,
		)
		_ = p.docRepo.MarkFailed(ctx, event.OrganizationID, event.DocumentID, "embed: "+err.Error())
		return fmt.Errorf("embed chunks: %w", err)
	}

	if len(vectors) != len(chunks) {
		_ = p.docRepo.MarkFailed(ctx, event.OrganizationID, event.DocumentID, "embed/chunk count mismatch")
		return fmt.Errorf("embed/chunk count mismatch: got=%d want=%d", len(vectors), len(chunks))
	}

	// 5. Ensure Qdrant collection exists (dimension from first vector)
	if err := p.vectorstore.EnsureCollection(ctx, uint64(len(vectors[0]))); err != nil {
		_ = p.docRepo.MarkFailed(ctx, event.OrganizationID, event.DocumentID, "qdrant ensure: "+err.Error())
		return fmt.Errorf("ensure qdrant collection: %w", err)
	}

	// 6. Upsert chunks with embeddings into Qdrant
	if err := p.vectorstore.ReplaceDocumentChunks(
		ctx,
		event.OrganizationID,
		event.DocumentID,
		event.ObjectKey,
		event.ContentType,
		chunks,
		vectors,
	); err != nil {
		p.logger.Error("Failed to upsert chunks into Qdrant",
			"doc_id", event.DocumentID,
			"org_id", event.OrganizationID,
			"error", err,
		)
		_ = p.docRepo.MarkFailed(ctx, event.OrganizationID, event.DocumentID, "qdrant upsert: "+err.Error())
		return fmt.Errorf("upsert chunks: %w", err)
	}

	// 7. Mark as indexed
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
		"chunks", len(chunks),
	)
	return nil
}
