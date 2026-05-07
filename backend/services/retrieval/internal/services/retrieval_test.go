package services

import (
	"context"
	"testing"

	tenantCtx "github.com/keshvan/rag-as-service/backend/pkg/common/grpc/context"
	"github.com/keshvan/rag-as-service/backend/services/retrieval/internal/qdrant"
)

func TestSearchRequiresOrganizationID(t *testing.T) {
	embedder := &fakeEmbedder{}
	searcher := &fakeSearcher{}
	service := NewRetrievalService(embedder, searcher, 5, 20, nil)

	_, err := service.Search(context.Background(), "hello", 5, 0)
	if err == nil {
		t.Fatal("expected error")
	}
	if err != ErrUnauthenticated {
		t.Fatalf("unexpected error: %v", err)
	}
	if embedder.called {
		t.Fatal("embedding should not be called without organization_id")
	}
	if searcher.called {
		t.Fatal("qdrant should not be called without organization_id")
	}
}

func TestSearchPassesOrganizationIDToVectorStore(t *testing.T) {
	embedder := &fakeEmbedder{vector: []float32{0.1, 0.2}}
	searcher := &fakeSearcher{
		points: []qdrant.ScoredPoint{{
			ID:    "point-1",
			Score: 0.9,
			Payload: map[string]any{
				"document_id":     "doc-1",
				"chunk_id":        "chunk-1",
				"text":            "hello",
				"organization_id": "org-1",
			},
		}},
	}
	service := NewRetrievalService(embedder, searcher, 5, 20, nil)

	ctx := tenantCtx.ToContext(context.Background(), "org-1")
	results, err := service.Search(ctx, "hello", 5, 0)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if searcher.orgID != "org-1" {
		t.Fatalf("unexpected org id: %s", searcher.orgID)
	}
	if len(results) != 1 {
		t.Fatalf("unexpected results count: %d", len(results))
	}
	if _, ok := results[0].Metadata["organization_id"]; ok {
		t.Fatal("organization_id should not be returned in metadata")
	}
}

type fakeEmbedder struct {
	called bool
	vector []float32
}

func (f *fakeEmbedder) EmbedOne(context.Context, string) ([]float32, error) {
	f.called = true
	if f.vector == nil {
		return []float32{0.1}, nil
	}
	return f.vector, nil
}

type fakeSearcher struct {
	called bool
	orgID  string
	points []qdrant.ScoredPoint
}

func (f *fakeSearcher) Search(_ context.Context, _ []float32, orgID string, _ int, _ *float32) ([]qdrant.ScoredPoint, error) {
	f.called = true
	f.orgID = orgID
	return f.points, nil
}
