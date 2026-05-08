package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/keshvan/rag-as-service/backend/pkg/common/embeddings"
	retrievalv1 "github.com/keshvan/rag-as-service/backend/pkg/common/gen/retrieval/v1"
	retClient "github.com/keshvan/rag-as-service/backend/services/gateway/internal/clients/retrieval"
)

type RetrievalHandler struct {
	retrievalClient *retClient.Client
	llmClient       *embeddings.YandexAIClient
	llmModel        string
}

func NewRetrievalHandler(client *retClient.Client, llmClient *embeddings.YandexAIClient, llmModel string) *RetrievalHandler {
	return &RetrievalHandler{
		retrievalClient: client,
		llmClient:       llmClient,
		llmModel:        llmModel,
	}
}

func (h *RetrievalHandler) Search(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query          string  `json:"query"`
		Limit          int32   `json:"limit"`
		ScoreThreshold float32 `json:"score_threshold"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if strings.TrimSpace(req.Query) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query is required"})
		return
	}

	resp, err := h.retrievalClient.Search(r.Context(), &retrievalv1.SearchRequest{
		Query:          req.Query,
		Limit:          req.Limit,
		ScoreThreshold: req.ScoreThreshold,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	results := make([]map[string]any, 0, len(resp.GetResults()))
	for _, item := range resp.GetResults() {
		results = append(results, map[string]any{
			"id":          item.GetId(),
			"document_id": item.GetDocumentId(),
			"chunk_id":    item.GetChunkId(),
			"text":        item.GetText(),
			"score":       item.GetScore(),
			"metadata":    item.GetMetadata(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *RetrievalHandler) RAGQuery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query          string  `json:"query"`
		Limit          int32   `json:"limit"`
		ScoreThreshold float32 `json:"score_threshold"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if strings.TrimSpace(req.Query) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query is required"})
		return
	}

	resp, err := h.retrievalClient.Search(r.Context(), &retrievalv1.SearchRequest{
		Query:          req.Query,
		Limit:          req.Limit,
		ScoreThreshold: req.ScoreThreshold,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	var contextParts []string
	for i, item := range resp.GetResults() {
		contextParts = append(contextParts, fmt.Sprintf("[Фрагмент %d]\n%s", i+1, item.GetText()))
	}

	systemPrompt := "Ты — полезный ассистент. Отвечай на вопрос пользователя, используя только предоставленный контекст из документов. Если ответ не найден в контексте, скажи об этом."
	userPrompt := fmt.Sprintf("Контекст:\n%s\n\nВопрос: %s", strings.Join(contextParts, "\n\n"), req.Query)

	answer, err := h.llmClient.Complete(r.Context(), h.llmModel, systemPrompt, userPrompt, 0.3, 2000)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "llm request failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"answer":  answer,
		"sources": resp.GetResults(),
	})
}
