package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/shared/logger"
)

type Connection struct {
	db     *sql.DB
	config config.DatabaseConfig
	log    *logger.Logger
}

func NewConnection(ctx context.Context, cfg config.DatabaseConfig, log *logger.Logger) (*Connection, error) {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	conn := &Connection{db: db, config: cfg, log: log}
	if err := conn.waitUntilReady(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return conn, nil
}

func (c *Connection) LogStatus(ctx context.Context) {
	stats := c.db.Stats()
	c.log.InfoContext(ctx, "database wiring ready", "driver", "postgres", "dsn_configured", c.config.DSN != "", "open_connections", stats.OpenConnections)
}

func (c *Connection) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, c.config.HealthCheckTimeout)
	defer cancel()

	return c.db.PingContext(pingCtx)
}

func (c *Connection) Close() error {
	if c.db == nil {
		return nil
	}

	return c.db.Close()
}

func (c *Connection) DB() *sql.DB {
	return c.db
}

func (c *Connection) waitUntilReady(ctx context.Context) error {
	deadlineCtx, cancel := context.WithTimeout(ctx, c.config.StartupPingTimeout)
	defer cancel()

	ticker := time.NewTicker(c.config.StartupPingInterval)
	defer ticker.Stop()

	for {
		if err := c.Ping(deadlineCtx); err == nil {
			return nil
		}

		select {
		case <-deadlineCtx.Done():
			if errors.Is(deadlineCtx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("ping postgres before startup: %w", deadlineCtx.Err())
			}
			return deadlineCtx.Err()
		case <-ticker.C:
		}
	}
}
