package app

import (
	"fmt"

	commonServer "github.com/keshvan/rag-as-service/backend/pkg/common/server"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/clients/auth"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/config"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/handlers"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/router"
)

type App struct {
	httpServer commonServer.Server
}

func New(cfg *config.GatewayConfig) (*App, error) {
	authGrpcClient, err := auth.New(cfg.AuthHost, cfg.AuthPort)
	if err != nil {
		return nil, fmt.Errorf("failed to init auth grpc client: %w", err)
	}

	authHandler := handlers.NewAuthHandler(authGrpcClient)

	appHandlers := router.Handlers{
		Auth: authHandler,
	}

	srv := commonServer.NewHTTPServer(cfg.BaseConfig.HTTP.Port)

	router.RegisterRoutes(srv.Router, appHandlers, cfg.JWTSecret)

	return &App{
		httpServer: *srv,
	}, nil
}

func (a *App) Run() error {
	return a.httpServer.Run()
}
