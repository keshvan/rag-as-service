package app

import (
	"log/slog"

	authv1 "github.com/keshvan/rag-as-service/backend/pkg/common/gen/auth/v1"
	commonGRPCServer "github.com/keshvan/rag-as-service/backend/pkg/common/grpc/server"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/cache"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/config"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/email"
	grpcserver "github.com/keshvan/rag-as-service/backend/services/auth/internal/grpc"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/repo"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/services"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/storage"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type App struct {
	grpcServer *commonGRPCServer.GRPCServer
	register   func(*grpc.Server)
}

func NewApp(cfg *config.Config) (*App, error) {
	logger := slog.Default()

	database, err := storage.InitDB(cfg)
	if err != nil {
		return nil, err
	}

	repository := repo.NewRepository(database)
	emailClient := buildEmailClient(cfg, logger)

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
	})
	redisStorage := cache.NewRedisVerificationStorage(redisClient)

	authService := services.NewAuth(logger, repository, repository, redisStorage, emailClient, cfg.Secret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	authGRPCServer := grpcserver.NewAuthServer(authService)
	grpcSrv := commonGRPCServer.NewGRPCServer(cfg.GRPC)

	register := func(server *grpc.Server) {
		authv1.RegisterAuthServiceServer(server, authGRPCServer)
	}

	//handler := handlers.NewAuthHandler(service)

	/*r := gin.Default()
	auth := r.Group("/auth")
	routes.RegisterRoutes(auth, handler, cfg.Secret)*/

	return &App{
		grpcServer: grpcSrv,
		register:   register,
	}, nil
}

func (a *App) Run() error {
	return a.grpcServer.Run(a.register)
}

func buildEmailClient(cfg *config.Config, logger *slog.Logger) services.EmailClient {
	if cfg.SMTP.User == "" || cfg.SMTP.Password == "" || cfg.SMTP.From == "" {
		logger.Warn("SMTP is not configured, verification codes will be logged only")
		return email.NewLoggingClient()
	}

	return email.NewSMTPClient(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.User, cfg.SMTP.Password, cfg.SMTP.From)
}
