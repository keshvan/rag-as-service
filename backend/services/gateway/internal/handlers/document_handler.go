package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	documentv1 "github.com/keshvan/rag-as-service/backend/pkg/common/gen/document/v1"
	docClient "github.com/keshvan/rag-as-service/backend/services/gateway/internal/clients/document"
)

type DocumentHandler struct {
	documentClient *docClient.Client
}

func NewDocumentHandler(client *docClient.Client) *DocumentHandler {
	return &DocumentHandler{documentClient: client}
}

func (h *DocumentHandler) InitUpload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FileName    string `json:"file_name"`
		ContentType string `json:"content_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	resp, err := h.documentClient.InitUpload(r.Context(), &documentv1.InitUploadRequest{
		FileName:    req.FileName,
		ContentType: req.ContentType,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"document_id": resp.GetDocumentId(),
		"upload_url":  resp.GetUploadUrl(),
		"headers":     resp.GetHeaders(),
		"object_key":  resp.GetObjectKey(),
		"expires_in":  resp.GetExpiresIn(),
	})
}

func (h *DocumentHandler) ConfirmUpload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DocumentID string `json:"document_id"`
		SizeBytes  int64  `json:"size_bytes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.DocumentID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.SizeBytes < 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	_, err := h.documentClient.CompleteUpload(r.Context(), &documentv1.CompleteUploadRequest{
		DocumentId: req.DocumentID,
		SizeBytes:  req.SizeBytes,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "uploaded"})
}

func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	limit := int32(0)
	offset := int32(0)

	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.ParseInt(rawLimit, 10, 32)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		limit = int32(parsedLimit)
	}

	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		parsedOffset, err := strconv.ParseInt(rawOffset, 10, 32)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		offset = int32(parsedOffset)
	}

	resp, err := h.documentClient.ListDocuments(r.Context(), &documentv1.ListDocumentsRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	items := make([]map[string]string, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		items = append(items, map[string]string{
			"id":           item.GetId(),
			"file_name":    item.GetFileName(),
			"content_type": item.GetContentType(),
			"status":       item.GetStatus(),
			"created_at":   item.GetCreatedAt(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
