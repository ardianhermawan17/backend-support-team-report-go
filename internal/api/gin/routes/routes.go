package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"backend-sport-team-report-go/internal/config"
	authhttp "backend-sport-team-report-go/internal/modules/auth/interfaces/http"
	playershttp "backend-sport-team-report-go/internal/modules/players/interfaces/http"
	teamshttp "backend-sport-team-report-go/internal/modules/teams/interfaces/http"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"
)

func Register(engine *gin.Engine, cfg config.Config, db *postgres.Connection, log *logger.Logger) {
	v1 := engine.Group("/api/v1")
	v1.GET("/health", func(c *gin.Context) {
		databaseStatus := "up"
		statusCode := http.StatusOK
		if db == nil {
			databaseStatus = "down"
			statusCode = http.StatusServiceUnavailable
			log.InfoContext(c.Request.Context(), "health check database unavailable", "path", c.FullPath(), "error", "database connection not configured")
		} else if err := db.Ping(c.Request.Context()); err != nil {
			databaseStatus = "down"
			statusCode = http.StatusServiceUnavailable
			log.InfoContext(c.Request.Context(), "health check database unavailable", "path", c.FullPath(), "error", err.Error())
		}

		log.InfoContext(c.Request.Context(), "health check", "path", c.FullPath())
		c.JSON(statusCode, gin.H{
			"status":     "ok",
			"service":    "soccer-team-report",
			"database":   databaseStatus,
			"checked_at": time.Now().UTC().Format(time.RFC3339),
		})
	})

	if db != nil {
		authMiddleware := authhttp.RegisterRoutes(v1, db, log, cfg.Auth)
		if err := teamshttp.RegisterRoutes(v1, db, log, authMiddleware); err != nil {
			log.InfoContext(context.Background(), "failed to register teams routes", "error", err.Error())
		}
		if err := playershttp.RegisterRoutes(v1, db, log, authMiddleware); err != nil {
			log.InfoContext(context.Background(), "failed to register players routes", "error", err.Error())
		}
	}
}
