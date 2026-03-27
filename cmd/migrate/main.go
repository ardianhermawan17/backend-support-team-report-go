package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"backend-sport-team-report-go/internal/config"
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

	files, err := filepath.Glob(filepath.Join("internal", "platform", "database", "migrations", "*.sql"))
	if err != nil {
		log.Fatalf("list migrations: %v", err)
	}
	sort.Strings(files)

	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("read migration %s: %v", file, err)
		}

		if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
			log.Fatalf("apply migration %s: %v", file, err)
		}

		fmt.Printf("applied migration %s\n", file)
	}
}
