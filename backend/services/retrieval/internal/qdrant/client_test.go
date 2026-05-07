package qdrant

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSearchAppliesOrganizationFilter(t *testing.T) {
	var got map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/collections/document_chunks/points/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":[{"id":"point-1","score":0.91,"payload":{"document_id":"doc-1","text":"hello"}}]}`))
	}))
	defer srv.Close()

	client, err := New(Config{
		URL:        srv.URL,
		Collection: "document_chunks",
		Timeout:    time.Second,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	points, err := client.Search(context.Background(), []float32{0.1, 0.2}, "org-1", 3, nil)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(points) != 1 {
		t.Fatalf("unexpected points count: %d", len(points))
	}

	filter, ok := got["filter"].(map[string]any)
	if !ok {
		t.Fatalf("missing filter: %#v", got["filter"])
	}
	must := filter["must"].([]any)
	cond := must[0].(map[string]any)
	if cond["key"] != "organization_id" {
		t.Fatalf("unexpected filter key: %v", cond["key"])
	}
	match := cond["match"].(map[string]any)
	if match["value"] != "org-1" {
		t.Fatalf("unexpected organization value: %v", match["value"])
	}
}
