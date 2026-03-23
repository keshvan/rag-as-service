package main

import (
	"log"

	app2 "github.com/keshvan/rag-as-service/backend/services/auth/internal/app"
	"github.com/keshvan/rag-as-service/backend/services/auth/internal/config"
)

func main() {
	cfg := config.Load()

	app, err := app2.NewApp(cfg)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	if err := app.Run(cfg.HTTP.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
