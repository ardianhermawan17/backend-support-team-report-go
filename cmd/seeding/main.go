package main

import (
	"context"
	"log"

	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/platform/database/seeding"
	"backend-sport-team-report-go/internal/shared/logger"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	appLogger := logger.New(cfg.App.Name, cfg.App.Env)
	conn, err := postgres.NewConnection(ctx, cfg.Database, appLogger)
	if err != nil {
		log.Fatalf("open database connection: %v", err)
	}
	defer conn.Close()

	service, err := seeding.NewService(conn.DB())
	if err != nil {
		log.Fatalf("create seeding service: %v", err)
	}

	if err := service.Seed(ctx); err != nil {
		log.Fatalf("seed database: %v", err)
	}

	log.Print("seeded database")
}
