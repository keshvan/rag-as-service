package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/lib/jwt"

	tenantCtx "github.com/keshvan/rag-as-service/backend/pkg/common/grpc/context"
)

type contextKey string

const (
	UserGUIDKey  contextKey = "user_guid"
	SessionIDKey contextKey = "session_id"
)

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Expose-Headers", "X-New-Access-Token, X-New-Refresh-Token")

			accessToken := extractTokenFromHeader(r.Header.Get("Authorization"))

			if accessToken == "" {
				http.Error(w, `{"error": "missing access token"}`, http.StatusUnauthorized)
				return
			}

			claims, err := jwt.ParseToken(accessToken, secret)
			if err != nil {
				http.Error(w, `{"error": "invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := r.Context()

			ctx = tenantCtx.ToContext(ctx, claims.OrganizationID)

			ctx = context.WithValue(ctx, UserGUIDKey, claims.UserGUID)
			ctx = context.WithValue(ctx, SessionIDKey, claims.SessionID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractTokenFromHeader(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}
