package main

import (
	"context"
	"log"

	"github.com/fieldsync/backend/internal/config"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/httpserver"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	router, err := httpserver.NewRouter(pool, cfg)
	if err != nil {
		log.Fatalf("setting up router: %v", err)
	}

	log.Printf("fieldsync api listening on :%s (env=%s)", cfg.Port, cfg.Env)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
