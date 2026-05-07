package main

import (
	"log"

	"github.com/keshvan/rag-as-service/backend/services/retrieval/internal/app"
	"github.com/keshvan/rag-as-service/backend/services/retrieval/internal/config"
)

func main() {
	cfg := config.MustLoad()

	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("failed to init retrieval app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("retrieval grpc server error: %v", err)
	}
}
