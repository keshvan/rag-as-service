package main

import (
	"context"
	"log"

	"github.com/keshvan/rag-as-service/backend/pkg/common/database"
	"github.com/keshvan/rag-as-service/backend/services/document/internal/app"
	"github.com/keshvan/rag-as-service/backend/services/document/internal/config"
)

func main() {
	cfg := config.MustLoad()

	if err := database.RunMigrations(
		context.Background(),
		cfg.Postgres.GetDSN(),
		"./migrations",
		database.MigrateOptions{},
	); err != nil {
		log.Fatalf("error while applying migrations: %v", err)
	}
	log.Println("migrations applied")

	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("failed to init document app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("document grpc server error: %v", err)
	}
}
