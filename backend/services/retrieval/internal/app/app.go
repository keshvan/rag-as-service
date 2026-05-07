package app

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/keshvan/rag-as-service/backend/pkg/common/embeddings"
	retrievalv1 "github.com/keshvan/rag-as-service/backend/pkg/common/gen/retrieval/v1"
	commonGRPCServer "github.com/keshvan/rag-as-service/backend/pkg/common/grpc/server"
	"github.com/keshvan/rag-as-service/backend/services/retrieval/internal/config"
	grpcserver "github.com/keshvan/rag-as-service/backend/services/retrieval/internal/grpc"
	"github.com/keshvan/rag-as-service/backend/services/retrieval/internal/qdrant"
	"github.com/keshvan/rag-as-service/backend/services/retrieval/internal/services"
	"google.golang.org/grpc"
)

type App struct {
	grpcServer *commonGRPCServer.GRPCServer
	register   func(*grpc.Server)
}

func NewApp(cfg *config.RetrievalConfig) (*App, error) {
	log := slog.Default()

	embedder, err := embeddings.NewYandexAIEmbedder(embeddings.YandexAIConfig{
		ApiKey:       cfg.YandexAPIKey,
		FolderID:     cfg.YandexFolderID,
		Model:        embeddingModel(cfg),
		BaseURL:      cfg.YandexEmbeddingBaseURL,
		MaxBatchSize: cfg.YandexEmbeddingMaxBatchSize,
	})
	if err != nil {
		return nil, fmt.Errorf("init yandex embedder: %w", err)
	}

	qdrantClient, err := qdrant.New(qdrant.Config{
		URL:        cfg.QdrantURL,
		APIKey:     cfg.QdrantAPIKey,
		Collection: cfg.QdrantCollection,
		VectorName: cfg.QdrantVectorName,
		Timeout:    time.Duration(cfg.QdrantTimeoutSeconds) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("init qdrant client: %w", err)
	}

	retrievalService := services.NewRetrievalService(
		embedder,
		qdrantClient,
		cfg.DefaultLimit,
		cfg.MaxLimit,
		log,
	)
	retrievalServer := grpcserver.NewServer(retrievalService)

	grpcSrv := commonGRPCServer.NewGRPCServer(cfg.GRPC)
	register := func(server *grpc.Server) {
		retrievalv1.RegisterRetrievalServiceServer(server, retrievalServer)
	}

	return &App{
		grpcServer: grpcSrv,
		register:   register,
	}, nil
}

func (a *App) Run() error {
	return a.grpcServer.Run(a.register)
}

func embeddingModel(cfg *config.RetrievalConfig) string {
	if cfg.YandexEmbeddingModel != "" {
		return cfg.YandexEmbeddingModel
	}
	return fmt.Sprintf("emb://%s/text-search-query/latest", cfg.YandexFolderID)
}
