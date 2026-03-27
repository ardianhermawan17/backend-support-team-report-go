package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/platform/database/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := sql.Open("pgx", cfg.Database.DSN)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := migrations.ApplyDir(ctx, db, "internal/platform/database/migrations"); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	fmt.Println("applied migrations")
}
