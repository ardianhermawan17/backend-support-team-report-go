package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"
)

func Register(engine *gin.Engine, db *postgres.Connection, log *logger.Logger) {
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
}
