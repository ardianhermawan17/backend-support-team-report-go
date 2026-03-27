package bootstrap

import (
	"context"

	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"
)

type Dependencies struct {
	Database *postgres.Connection
}

func NewDependencies(ctx context.Context, cfg config.Config, log *logger.Logger) (*Dependencies, error) {
	database, err := postgres.NewConnection(ctx, cfg.Database, log)
	if err != nil {
		return nil, err
	}

	return &Dependencies{
		Database: database,
	}, nil
}
