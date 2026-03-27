package testenv

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/platform/database/migrations"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"
)

type PostgresEnv struct {
	DSN string
	DB  *sql.DB
}

func StartPostgres(t *testing.T) *PostgresEnv {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:17-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       "soccer_team_report_test",
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(2 * time.Minute),
		},
	})
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("postgres host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("postgres mapped port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/soccer_team_report_test?sslmode=disable", host, port.Port())

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open postgres test db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := waitForDatabase(ctx, db); err != nil {
		t.Fatalf("wait for postgres ping: %v", err)
	}

	migrationsDir := filepath.Join(repoRoot(t), "internal", "platform", "database", "migrations")
	if err := migrations.ApplyDir(ctx, db, migrationsDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	return &PostgresEnv{DSN: dsn, DB: db}
}

func (e *PostgresEnv) DatabaseConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		DSN:                 e.DSN,
		MaxOpenConns:        5,
		MaxIdleConns:        5,
		ConnMaxLifetime:     time.Minute,
		HealthCheckTimeout:  2 * time.Second,
		StartupPingTimeout:  15 * time.Second,
		StartupPingInterval: 250 * time.Millisecond,
	}
}

func (e *PostgresEnv) OpenConnection(t *testing.T) *postgres.Connection {
	t.Helper()

	conn, err := postgres.NewConnection(context.Background(), e.DatabaseConfig(), logger.New("integration-test", config.EnvTest))
	if err != nil {
		t.Fatalf("open postgres connection: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
	})

	return conn
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve repo root")
	}

	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
	if root == "." {
		t.Fatal(fmt.Errorf("invalid repo root resolved"))
	}

	return root
}

func waitForDatabase(ctx context.Context, db *sql.DB) error {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		if err := db.PingContext(ctx); err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
