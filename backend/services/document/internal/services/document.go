package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	tenantCtx "github.com/keshvan/rag-as-service/backend/pkg/common/grpc/context"
	"github.com/keshvan/rag-as-service/backend/services/document/internal/config"
	"github.com/keshvan/rag-as-service/backend/services/document/internal/entity"
	"github.com/keshvan/rag-as-service/backend/services/document/internal/events"
)

var (
	ErrUnauthenticated = errors.New("missing organization context")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotFound        = errors.New("document not found")
)

type Repository interface {
	CreatePending(ctx context.Context, d entity.Document) error
	MarkUploaded(ctx context.Context, orgID, documentID string, sizeBytes int64) error
	ListByOrg(ctx context.Context, orgID string, limit, offset int) ([]entity.Document, error)
	GetDocumentByID(ctx context.Context, orgID, documentID string) (*entity.Document, error)
}

type PresignService interface {
	PresignPut(ctx context.Context, objectKey, contentType string) (string, map[string]string, int32, error)
	BuildObjectKey(orgID, documentID, fileName string) string
}

type EventProducer interface {
	PublishDocumentUploaded(ctx context.Context, event events.DocumentUploadedEvent) error
}

type DocumentService struct {
	repo      Repository
	s3        PresignService
	cfg       *config.DocumentConfig
	producer  EventProducer
	logger    *slog.Logger
}

type InitUploadOutput struct {
	DocumentID string
	UploadURL  string
	Headers    map[string]string
	ObjectKey  string
	ExpiresIn  int32
}

type ListItem struct {
	ID          string
	FileName    string
	ContentType string
	Status      entity.DocumentStatus
	CreatedAt   string
}

func NewDocumentService(r Repository, s3 PresignService, cfg *config.DocumentConfig, producer EventProducer, logger *slog.Logger) *DocumentService {
	return &DocumentService{repo: r, s3: s3, cfg: cfg, producer: producer, logger: logger}
}

func (s *DocumentService) InitUpload(ctx context.Context, fileName, contentType string) (*InitUploadOutput, error) {
	orgID, ok := tenantCtx.FromContext(ctx)
	if !ok || orgID == "" {
		return nil, ErrUnauthenticated
	}

	fileName = strings.TrimSpace(fileName)
	contentType = strings.TrimSpace(contentType)
	if fileName == "" || contentType == "" {
		return nil, ErrInvalidArgument
	}

	docID := uuid.NewString()
	objectKey := s.s3.BuildObjectKey(orgID, docID, fileName)

	if err := s.repo.CreatePending(ctx, entity.Document{
		ID:          docID,
		OrgID:       orgID,
		FileName:    fileName,
		ContentType: contentType,
		Status:      entity.DocumentStatusPendingUpload,
		ObjectKey:   objectKey,
		SizeBytes:   0,
	}); err != nil {
		return nil, err
	}

	url, headers, expiresIn, err := s.s3.PresignPut(ctx, objectKey, contentType)
	if err != nil {
		return nil, err
	}

	return &InitUploadOutput{
		DocumentID: docID,
		UploadURL:  url,
		Headers:    headers,
		ObjectKey:  objectKey,
		ExpiresIn:  expiresIn,
	}, nil
}

func (s *DocumentService) CompleteUpload(ctx context.Context, documentID string, sizeBytes int64) error {
	orgID, ok := tenantCtx.FromContext(ctx)
	if !ok || orgID == "" {
		return ErrUnauthenticated
	}
	if strings.TrimSpace(documentID) == "" {
		return ErrInvalidArgument
	}
	if sizeBytes < 0 {
		return ErrInvalidArgument
	}

	// Get document details before marking as uploaded
	doc, err := s.repo.GetDocumentByID(ctx, orgID, documentID)
	if err != nil {
		return ErrNotFound
	}

	if err := s.repo.MarkUploaded(ctx, orgID, documentID, sizeBytes); err != nil {
		return ErrNotFound
	}

	// Publish event to Kafka (fire-and-forget with logging)
	if s.producer != nil {
		publishCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		event := events.DocumentUploadedEvent{
			EventID:        uuid.NewString(),
			EventType:      "DOCUMENT_UPLOADED",
			DocumentID:     documentID,
			OrganizationID: orgID,
			ObjectKey:      doc.ObjectKey,
			ContentType:    doc.ContentType,
			SizeBytes:      sizeBytes,
			OccurredAt:     time.Now().UTC().Format(time.RFC3339),
		}

		if err := s.producer.PublishDocumentUploaded(publishCtx, event); err != nil {
			// Log but don't fail the upload completion
			s.logger.Error("Failed to publish document uploaded event",
				"document_id", documentID,
				"organization_id", orgID,
				"error", err,
			)
		}
	}

	return nil
}

func (s *DocumentService) ListDocuments(ctx context.Context, limit, offset int) ([]ListItem, error) {
	orgID, ok := tenantCtx.FromContext(ctx)
	if !ok || orgID == "" {
		return nil, ErrUnauthenticated
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	docs, err := s.repo.ListByOrg(ctx, orgID, limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]ListItem, 0, len(docs))
	for _, d := range docs {
		out = append(out, ListItem{
			ID:          d.ID,
			FileName:    d.FileName,
			ContentType: d.ContentType,
			Status:      d.Status,
			CreatedAt:   d.CreatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}
