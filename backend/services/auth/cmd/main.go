package main

import (
	"log"

	"github.com/keshvan/rag-as-service/backend/services/auth/internal/app"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/config"
)

func main() {
	cfg := config.Load()

	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("grpc server error: %v", err)
	}
}
