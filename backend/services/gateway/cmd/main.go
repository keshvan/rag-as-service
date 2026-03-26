package main

import (
	"log"
	"net/http"

	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/app"
	"github.com/keshvan/rag-as-service/backend/services/gateway/internal/config"
)

func main() {
	cfg := config.MustLoad()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	log.Printf("Starting API Gateway on port :%d", cfg.HTTP.Port)

	if err := application.Run(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Application failed: %v", err)
	}

	log.Println("Application successfully stopped")
}
