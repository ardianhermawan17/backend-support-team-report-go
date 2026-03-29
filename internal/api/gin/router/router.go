package router

import (
	"context"
	"net/http"
	"time"

	authhttp "backend-sport-team-report-go/internal/modules/auth/interfaces/http"
	playershttp "backend-sport-team-report-go/internal/modules/player/interfaces/http"
	reportshttp "backend-sport-team-report-go/internal/modules/report/interfaces/http"
	scheduleshttp "backend-sport-team-report-go/internal/modules/schedule/interfaces/http"
	teamshttp "backend-sport-team-report-go/internal/modules/team/interfaces/http"
	"github.com/gin-gonic/gin"

	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"
	sharedmiddleware "backend-sport-team-report-go/internal/shared/middleware"
)

func New(cfg config.Config, db *postgres.Connection, log *logger.Logger) *gin.Engine {
	if cfg.App.Env == config.EnvProd {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.HandleMethodNotAllowed = true
	engine.Use(gin.Recovery())
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "route not found"})
	})
	engine.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method_not_allowed", "message": "method is not allowed for this route"})
	})

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
		loginRateLimitMiddleware := sharedmiddleware.NewRateLimitMiddleware(
			log,
			"auth_login",
			cfg.Security.RateLimit.Login.MaxRequests,
			cfg.Security.RateLimit.Login.Window,
			sharedmiddleware.ClientRouteKey,
		)
		writeRateLimitMiddleware := sharedmiddleware.NewRateLimitMiddleware(
			log,
			"authenticated_write",
			cfg.Security.RateLimit.AuthenticatedWrite.MaxRequests,
			cfg.Security.RateLimit.AuthenticatedWrite.Window,
			sharedmiddleware.AuthenticatedRouteKey,
		)

		authMiddleware := authhttp.RegisterRoutes(v1, db, log, cfg.Auth, cfg.Security, loginRateLimitMiddleware)
		if err := teamshttp.RegisterRoutes(v1, db, log, authMiddleware, cfg.Security, writeRateLimitMiddleware); err != nil {
			log.InfoContext(context.Background(), "failed to register teams routes", "error", err.Error())
		}
		if err := playershttp.RegisterRoutes(v1, db, log, authMiddleware, cfg.Security, writeRateLimitMiddleware); err != nil {
			log.InfoContext(context.Background(), "failed to register players routes", "error", err.Error())
		}
		if err := scheduleshttp.RegisterRoutes(v1, db, log, authMiddleware, cfg.Security, writeRateLimitMiddleware); err != nil {
			log.InfoContext(context.Background(), "failed to register schedules routes", "error", err.Error())
		}
		if err := reportshttp.RegisterRoutes(v1, db, log, authMiddleware, cfg.Security, writeRateLimitMiddleware); err != nil {
			log.InfoContext(context.Background(), "failed to register reports routes", "error", err.Error())
		}
	}

	return engine
}
