package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	tenantCtx "github.com/keshvan/rag-as-service/backend/pkg/common/grpc/context"
	"github.com/keshvan/rag-as-service/backend/services/retrieval/internal/qdrant"
)

var (
	ErrUnauthenticated = errors.New("organization_id is required")
	ErrInvalidArgument = errors.New("invalid argument")
)

type Embedder interface {
	EmbedOne(ctx context.Context, text string) ([]float32, error)
}

type VectorSearcher interface {
	Search(ctx context.Context, vector []float32, orgID string, limit int, scoreThreshold *float32) ([]qdrant.ScoredPoint, error)
}

type RetrievalService struct {
	embedder     Embedder
	searcher     VectorSearcher
	defaultLimit int
	maxLimit     int
	log          *slog.Logger
}

type SearchResult struct {
	ID         string
	DocumentID string
	ChunkID    string
	Text       string
	Score      float32
	Metadata   map[string]string
}

func NewRetrievalService(embedder Embedder, searcher VectorSearcher, defaultLimit, maxLimit int, log *slog.Logger) *RetrievalService {
	if defaultLimit <= 0 {
		defaultLimit = 5
	}
	if maxLimit <= 0 {
		maxLimit = 20
	}
	if defaultLimit > maxLimit {
		defaultLimit = maxLimit
	}

	return &RetrievalService{
		embedder:     embedder,
		searcher:     searcher,
		defaultLimit: defaultLimit,
		maxLimit:     maxLimit,
		log:          log,
	}
}

func (s *RetrievalService) Search(ctx context.Context, query string, limit int, scoreThreshold float32) ([]SearchResult, error) {
	orgID, ok := tenantCtx.FromContext(ctx)
	if !ok || strings.TrimSpace(orgID) == "" {
		return nil, ErrUnauthenticated
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("%w: query is required", ErrInvalidArgument)
	}

	limit = s.normalizeLimit(limit)

	vector, err := s.embedder.EmbedOne(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get query embedding: %w", err)
	}

	var threshold *float32
	if scoreThreshold > 0 {
		threshold = &scoreThreshold
	}

	points, err := s.searcher.Search(ctx, vector, orgID, limit, threshold)
	if err != nil {
		return nil, fmt.Errorf("search qdrant: %w", err)
	}

	results := make([]SearchResult, 0, len(points))
	for _, point := range points {
		results = append(results, SearchResult{
			ID:         point.ID,
			DocumentID: payloadString(point.Payload, "document_id", "documentId", "doc_id"),
			ChunkID:    payloadString(point.Payload, "chunk_id", "chunkId"),
			Text:       payloadString(point.Payload, "text", "content", "chunk_text"),
			Score:      point.Score,
			Metadata:   payloadMetadata(point.Payload),
		})
	}

	return results, nil
}

func (s *RetrievalService) normalizeLimit(limit int) int {
	if limit <= 0 {
		return s.defaultLimit
	}
	if limit > s.maxLimit {
		return s.maxLimit
	}
	return limit
}

func payloadString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}
		switch v := value.(type) {
		case string:
			return v
		case fmt.Stringer:
			return v.String()
		default:
			return fmt.Sprint(v)
		}
	}
	return ""
}

func payloadMetadata(payload map[string]any) map[string]string {
	metadata := make(map[string]string, len(payload))
	for key, value := range payload {
		switch key {
		case "organization_id", "text", "content", "chunk_text":
			continue
		}

		switch v := value.(type) {
		case string:
			metadata[key] = v
		case float64, bool:
			metadata[key] = fmt.Sprint(v)
		default:
			raw, err := json.Marshal(v)
			if err != nil {
				metadata[key] = fmt.Sprint(v)
				continue
			}
			metadata[key] = string(raw)
		}
	}
	return metadata
}
