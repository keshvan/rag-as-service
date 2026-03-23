package routes

import (
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/handlers"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *handlers.AuthHandler, secret string) {
	// public routes
	rg.POST("/register", handler.SendEmailWithCode)
	rg.POST("/verification", handler.VerifyEmail)
	rg.POST("/login", handler.Login)

	// private routes
	protected := rg.Group("/")
	protected.Use(middleware.AuthMiddleware(secret))

	protected.GET("/me", handler.GetCurrentUserGUID)
	protected.GET("/tokens", handler.GetTokenPairByUserGUID)
	protected.POST("/logout", handler.Logout)
	protected.POST("/refresh", handler.RefreshTokens)
}
