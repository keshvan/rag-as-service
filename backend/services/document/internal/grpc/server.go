package grpcserver

import (
	"context"

	documentv1 "github.com/keshvan/rag-as-service/backend/pkg/common/gen/document/v1"
	"github.com/keshvan/rag-as-service/backend/services/document/internal/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DocumentService interface {
	InitUpload(ctx context.Context, fileName, contentType string) (*services.InitUploadOutput, error)
	CompleteUpload(ctx context.Context, documentID string, sizeBytes int64) error
	ListDocuments(ctx context.Context, limit, offset int) ([]services.ListItem, error)
}

type DocumentServer struct {
	documentv1.UnimplementedDocumentServiceServer
	doc DocumentService
}

func NewServer(doc DocumentService) *DocumentServer {
	return &DocumentServer{doc: doc}
}

func (s *DocumentServer) InitUpload(ctx context.Context, req *documentv1.InitUploadRequest) (*documentv1.InitUploadResponse, error) {
	out, err := s.doc.InitUpload(ctx, req.GetFileName(), req.GetContentType())
	if err != nil {
		return nil, mapError(err)
	}

	return &documentv1.InitUploadResponse{
		DocumentId: out.DocumentID,
		UploadUrl:  out.UploadURL,
		Headers:    out.Headers,
		ObjectKey:  out.ObjectKey,
		ExpiresIn:  out.ExpiresIn,
	}, nil
}

func (s *DocumentServer) CompleteUpload(ctx context.Context, req *documentv1.CompleteUploadRequest) (*documentv1.CompleteUploadResponse, error) {
	if err := s.doc.CompleteUpload(ctx, req.GetDocumentId(), req.GetSizeBytes()); err != nil {
		return nil, mapError(err)
	}
	return &documentv1.CompleteUploadResponse{Status: "uploaded"}, nil
}

func (s *DocumentServer) ListDocuments(ctx context.Context, req *documentv1.ListDocumentsRequest) (*documentv1.ListDocumentsResponse, error) {
	items, err := s.doc.ListDocuments(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapError(err)
	}

	resp := &documentv1.ListDocumentsResponse{Items: make([]*documentv1.DocumentItem, 0, len(items))}
	for _, it := range items {
		resp.Items = append(resp.Items, &documentv1.DocumentItem{
			Id:          it.ID,
			FileName:    it.FileName,
			ContentType: it.ContentType,
			Status:      string(it.Status),
			CreatedAt:   it.CreatedAt,
		})
	}
	return resp, nil
}

func mapError(err error) error {
	switch {
	case err == services.ErrUnauthenticated:
		return status.Error(codes.Unauthenticated, err.Error())
	case err == services.ErrInvalidArgument:
		return status.Error(codes.InvalidArgument, err.Error())
	case err == services.ErrNotFound:
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
