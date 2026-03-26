package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/handlers"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/middleware"
)

type Handlers struct {
	Auth *handlers.AuthHandler
	// Document
	// RAG
	// ...
}

func RegisterRoutes(r *chi.Mux, h Handlers, jwtSecret string) {
	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Route("/auth", func(authRouter chi.Router) {
			authRouter.Post("/login", h.Auth.Login)
			authRouter.Post("/register", h.Auth.Register)
			authRouter.Post("/verification", h.Auth.Verification)
		})

		v1.Group(func(protected chi.Router) {
			protected.Use(middleware.AuthMiddleware(jwtSecret))

			protected.Get("/me", h.Auth.GetCurrentUserGUID)
			protected.Get("/tokens", h.Auth.GetTokenPairByUserGUID)
			protected.Post("/logout", h.Auth.Logout)
			protected.Post("/refresh", h.Auth.RefreshTokens)
		})
	})
}
