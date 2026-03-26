package app

import (
	"fmt"
	"log/slog"

	commonLogger "github.com/keshvan/rag-as-service/backend/pkg/common/logger"
	commonServer "github.com/keshvan/rag-as-service/backend/pkg/common/server"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/clients/auth"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/config"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/handlers"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/router"
)

type App struct {
	httpServer commonServer.Server
	authClient *auth.Client
	log        *slog.Logger
}

func New(cfg *config.GatewayConfig) (*App, error) {
	log := commonLogger.New(cfg.AppEnv)

	authClient, err := auth.New(cfg.AuthHost, cfg.AuthPort)
	if err != nil {
		return nil, fmt.Errorf("failed to init auth grpc client: %w", err)
	}

	authHandler := handlers.NewAuthHandler(authClient)

	appHandlers := router.Handlers{
		Auth: authHandler,
	}

	srv := commonServer.NewHTTPServer(
		cfg.HTTP.Host,
		cfg.HTTP.Port,
		log,
	)

	router.RegisterRoutes(srv.Router, appHandlers, cfg.JWTSecret)

	return &App{
		httpServer: *srv,
		authClient: authClient,
		log:        log,
	}, nil
}

func (a *App) Run() error {
	defer func() {
		if err := a.authClient.Close(); err != nil {
			a.log.Error("failed to close auth client", "err", err)
		}
	}()

	return a.httpServer.Run()
}
