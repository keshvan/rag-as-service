package app

import (
	"fmt"
	"log/slog"

	"github.com/keshvan/rag-as-service/backend/pkg/common/embeddings"
	commonLogger "github.com/keshvan/rag-as-service/backend/pkg/common/logger"
	commonServer "github.com/keshvan/rag-as-service/backend/pkg/common/server"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/clients/auth"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/clients/document"
	retClient "github.com/keshvan/rag-as-service/backend/services/gateway/internal/clients/retrieval"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/config"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/handlers"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/router"
)

type App struct {
	httpServer      commonServer.Server
	authClient      *auth.Client
	retrievalClient *retClient.Client
	log             *slog.Logger
}

func New(cfg *config.GatewayConfig) (*App, error) {
	log := commonLogger.New(cfg.AppEnv)

	authClient, err := auth.New(cfg.AuthHost, cfg.AuthPort)
	if err != nil {
		return nil, fmt.Errorf("failed to init auth grpc client: %w", err)
	}

	documentClient, err := document.New(cfg.DocumentHost, cfg.DocumentPort)
	if err != nil {
		return nil, fmt.Errorf("failed to init document grpc client: %w", err)
	}

	retrievalClient, err := retClient.New(cfg.RetrievalHost, cfg.RetrievalPort)
	if err != nil {
		return nil, fmt.Errorf("failed to init retrieval grpc client: %w", err)
	}

	llmClient, err := embeddings.NewYandexAIClient(embeddings.YandexAIConfig{
		ApiKey:   cfg.YandexAPIKey,
		FolderID: cfg.YandexFolderID,
		BaseURL:  cfg.YandexLLMBaseURL,
		Model:    "", // embeddings model not needed for LLM, Complete uses its own model
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init yandex llm client: %w", err)
	}

	authHandler := handlers.NewAuthHandler(authClient)
	documentHandler := handlers.NewDocumentHandler(documentClient)
	retrievalHandler := handlers.NewRetrievalHandler(retrievalClient, llmClient, cfg.YandexLLMModel)

	appHandlers := router.Handlers{
		Auth:     authHandler,
		Document: documentHandler,
		Retrieval: retrievalHandler,
	}

	srv := commonServer.NewHTTPServer(
		cfg.HTTP.Host,
		cfg.HTTP.Port,
		log,
	)

	router.RegisterRoutes(srv.Router, appHandlers, cfg.JWTSecret)

	return &App{
		httpServer:      *srv,
		authClient:      authClient,
		retrievalClient: retrievalClient,
		log:             log,
	}, nil
}

func (a *App) Run() error {
	defer func() {
		if err := a.authClient.Close(); err != nil {
			a.log.Error("failed to close auth client", "err", err)
		}
		if err := a.retrievalClient.Close(); err != nil {
			a.log.Error("failed to close retrieval client", "err", err)
		}
	}()

	return a.httpServer.Run()
}
