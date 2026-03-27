package router

import (
	"github.com/gin-gonic/gin"

	ginroutes "backend-sport-team-report-go/internal/api/gin/routes"
	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"
)

func New(cfg config.Config, db *postgres.Connection, log *logger.Logger) *gin.Engine {
	if cfg.App.Env == config.EnvProd {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	ginroutes.Register(engine, db, log)
	return engine
}
