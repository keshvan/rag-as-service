package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/keshvan/rag-as-service/backend/pkg/common/database"
	documentv1 "github.com/keshvan/rag-as-service/backend/pkg/common/gen/document/v1"
	commonGRPCServer "github.com/keshvan/rag-as-service/backend/pkg/common/grpc/server"
	"github.com/keshvan/rag-as-service/backend/services/document/internal/config"
	grpcserver "github.com/keshvan/rag-as-service/backend/services/document/internal/grpc"
	"github.com/keshvan/rag-as-service/backend/services/document/internal/repo"
	"github.com/keshvan/rag-as-service/backend/services/document/internal/services"
	"google.golang.org/grpc"
)

type App struct {
	grpcServer *commonGRPCServer.GRPCServer
	register   func(*grpc.Server)
}

func NewApp(cfg *config.DocumentConfig) (*App, error) {
	log := slog.Default()

	pool, err := database.NewPostgresPool(context.Background(), cfg.Postgres.GetDSN(), database.DBOptions{
		MaxRetries:     15,
		RetryInterval:  2 * time.Second,
		MaxConnections: 20,
	}, log)
	if err != nil {
		return nil, fmt.Errorf("init postgres pool: %w", err)
	}

	repository := repo.NewDocumentRepository(pool)
	s3Service, err := services.NewS3PresignService(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("init s3 presign service: %w", err)
	}

	documentService := services.NewDocumentService(repository, s3Service, cfg)
	documentServer := grpcserver.NewServer(documentService)

	grpcSrv := commonGRPCServer.NewGRPCServer(cfg.GRPC)
	register := func(server *grpc.Server) {
		documentv1.RegisterDocumentServiceServer(server, documentServer)
	}

	return &App{
		grpcServer: grpcSrv,
		register:   register,
	}, nil
}

func (a *App) Run() error {
	return a.grpcServer.Run(a.register)
}
