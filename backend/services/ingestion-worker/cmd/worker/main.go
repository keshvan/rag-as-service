package main

import (
	"log"

	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/app"
	"github.com/keshvan/rag-as-service/backend/services/ingestion-worker/internal/config"
)

func main() {
	cfg := config.LoadConfig()

	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("failed to init ingestion worker app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("ingestion worker app error: %v", err)
	}
}
