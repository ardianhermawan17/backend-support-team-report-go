package http

import (
	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/modules/player/application"
	playersid "backend-sport-team-report-go/internal/modules/player/infrastructure/id"
	playerspersistence "backend-sport-team-report-go/internal/modules/player/infrastructure/persistence"
	interfacehandlers "backend-sport-team-report-go/internal/modules/player/interfaces/http/handlers"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(v1 *gin.RouterGroup, db *postgres.Connection, log *logger.Logger, authMiddleware gin.HandlerFunc, security config.SecurityConfig, writeRateLimitMiddleware gin.HandlerFunc) error {
	repository := playerspersistence.NewPlayerRepository(db)
	idGenerator, err := playersid.NewSnowflakeGenerator(2)
	if err != nil {
		return err
	}
	service := application.NewService(repository, idGenerator)
	handler := interfacehandlers.NewHandler(log, service, security.MaxJSONBodyBytes)

	playersGroup := v1.Group("/teams/:team_id/players")
	playersGroup.Use(authMiddleware)
	playersGroup.POST("", writeRateLimitMiddleware, handler.Create)
	playersGroup.GET("", handler.List)
	playersGroup.GET("/:player_id", handler.GetByID)
	playersGroup.PUT("/:player_id", writeRateLimitMiddleware, handler.Update)
	playersGroup.DELETE("/:player_id", writeRateLimitMiddleware, handler.Delete)

	return nil
}
