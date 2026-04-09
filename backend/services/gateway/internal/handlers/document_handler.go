package handlers

import (
	"encoding/json"
	"net/http"

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
