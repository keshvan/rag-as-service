package grpcserver

import (
	"context"
	"errors"

	retrievalv1 "github.com/keshvan/rag-as-service/backend/pkg/common/gen/retrieval/v1"
	"github.com/keshvan/rag-as-service/backend/services/retrieval/internal/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RetrievalService interface {
	Search(ctx context.Context, query string, limit int, scoreThreshold float32) ([]services.SearchResult, error)
}

type RetrievalServer struct {
	retrievalv1.UnimplementedRetrievalServiceServer
	retrieval RetrievalService
}

func NewServer(retrieval RetrievalService) *RetrievalServer {
	return &RetrievalServer{retrieval: retrieval}
}

func (s *RetrievalServer) Search(ctx context.Context, req *retrievalv1.SearchRequest) (*retrievalv1.SearchResponse, error) {
	items, err := s.retrieval.Search(ctx, req.GetQuery(), int(req.GetLimit()), req.GetScoreThreshold())
	if err != nil {
		return nil, mapError(err)
	}

	resp := &retrievalv1.SearchResponse{Results: make([]*retrievalv1.SearchResult, 0, len(items))}
	for _, item := range items {
		resp.Results = append(resp.Results, &retrievalv1.SearchResult{
			Id:         item.ID,
			DocumentId: item.DocumentID,
			ChunkId:    item.ChunkID,
			Text:       item.Text,
			Score:      item.Score,
			Metadata:   item.Metadata,
		})
	}

	return resp, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, services.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, services.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
