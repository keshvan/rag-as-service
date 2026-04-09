package handlers

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	authv1 "github.com/keshvan/rag-as-service/backend/pkg/common/gen/auth/v1"
	authClient "github.com/keshvan/rag-as-service/backend/services/gateway/internal/clients/auth"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authClient *authClient.Client
}

func NewAuthHandler(client *authClient.Client) *AuthHandler {
	return &AuthHandler{
		authClient: client,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	resp, err := h.authClient.Login(r.Context(), &authv1.LoginRequest{
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: r.UserAgent(),
		Ip:        clientIP(r),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"guid":          resp.GetGuid(),
		"access_token":  resp.GetAccessToken(),
		"refresh_token": resp.GetRefreshToken(),
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		OrganizationName string `json:"organization_name"`
		OrganizationURL  string `json:"organization_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	resp, err := h.authClient.Register(r.Context(), &authv1.RegisterRequest{
		Email:            req.Email,
		Password:         req.Password,
		OrganizationName: req.OrganizationName,
		OrganizationUrl:  req.OrganizationURL,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": resp.GetMessage(),
	})
}

func (h *AuthHandler) Verification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	resp, err := h.authClient.VerifyEmail(r.Context(), &authv1.VerifyEmailRequest{
		Email: req.Email,
		Code:  req.Code,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"guid": resp.GetGuid(),
	})
}

func (h *AuthHandler) GetCurrentUserGUID(w http.ResponseWriter, r *http.Request) {
	userGUID, ok := userGUIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing user context"})
		return
	}

	resp, err := h.authClient.GetCurrentUserGUID(r.Context(), &authv1.GetCurrentUserGUIDRequest{
		UserGuid: userGUID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"guid": resp.GetGuid(),
	})
}

func (h *AuthHandler) GetTokenPairByUserGUID(w http.ResponseWriter, r *http.Request) {
	userGUID, ok := userGUIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing user context"})
		return
	}

	var req struct {
		GUID string `json:"guid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	resp, err := h.authClient.GetTokenPairByUserGUID(r.Context(), &authv1.GetTokenPairByUserGUIDRequest{
		CurrentUserGuid: userGUID,
		TargetGuid:      req.GUID,
		UserAgent:       r.UserAgent(),
		Ip:              clientIP(r),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  resp.GetAccessToken(),
		"refresh_token": resp.GetRefreshToken(),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userGUID, ok := userGUIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing user context"})
		return
	}

	_, err := h.authClient.Logout(r.Context(), &authv1.LogoutRequest{
		UserGuid: userGUID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	userGUID, ok := userGUIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing user context"})
		return
	}

	sessionID, ok := sessionIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing session context"})
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	resp, err := h.authClient.RefreshTokens(r.Context(), &authv1.RefreshTokensRequest{
		UserGuid:     userGUID,
		SessionId:    sessionID,
		RefreshToken: req.RefreshToken,
		UserAgent:    r.UserAgent(),
		Ip:           clientIP(r),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  resp.GetAccessToken(),
		"refresh_token": resp.GetRefreshToken(),
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	httpCode := http.StatusInternalServerError
	switch st.Code() {
	case codes.InvalidArgument:
		httpCode = http.StatusBadRequest
	case codes.Unauthenticated:
		httpCode = http.StatusUnauthorized
	case codes.PermissionDenied:
		httpCode = http.StatusForbidden
	case codes.NotFound:
		httpCode = http.StatusNotFound
	case codes.AlreadyExists:
		httpCode = http.StatusConflict
	case codes.Internal:
		httpCode = http.StatusInternalServerError
	}

	writeJSON(w, httpCode, map[string]string{"error": st.Message()})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}

func userGUIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(middleware.UserGUIDKey).(string)
	return v, ok
}

func sessionIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(middleware.SessionIDKey).(string)
	return v, ok
}
