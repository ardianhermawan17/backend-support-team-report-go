package bootstrap

import (
	"context"
	"time"

	ginrouter "backend-sport-team-report-go/internal/api/gin/router"
	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/platform/httpserver"
	"backend-sport-team-report-go/internal/shared/logger"
)

type App struct {
	config       config.Config
	dependencies *Dependencies
	server       *httpserver.Server
	log          *logger.Logger
}

func NewApp(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	appLogger := logger.New(cfg.App.Name, cfg.App.Env)
	deps, err := NewDependencies(ctx, cfg, appLogger)
	if err != nil {
		return nil, err
	}

	router := ginrouter.New(cfg, deps.Database, appLogger)
	server := httpserver.New(cfg.App.Address(), cfg.App.ReadTimeout, cfg.App.WriteTimeout, router)

	return &App{
		config:       cfg,
		dependencies: deps,
		server:       server,
		log:          appLogger,
	}, nil
}

func (a *App) ShutdownTimeout() time.Duration {
	return a.config.App.ShutdownTimeout
}
