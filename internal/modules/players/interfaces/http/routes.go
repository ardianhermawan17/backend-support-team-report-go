package http

import (
	"backend-sport-team-report-go/internal/modules/players/application"
	playersid "backend-sport-team-report-go/internal/modules/players/infrastructure/id"
	playerspersistence "backend-sport-team-report-go/internal/modules/players/infrastructure/persistence"
	interfacehandlers "backend-sport-team-report-go/internal/modules/players/interfaces/http/handlers"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(v1 *gin.RouterGroup, db *postgres.Connection, log *logger.Logger, authMiddleware gin.HandlerFunc) error {
	repository := playerspersistence.NewPlayerRepository(db)
	idGenerator, err := playersid.NewSnowflakeGenerator(2)
	if err != nil {
		return err
	}
	service := application.NewService(repository, idGenerator)
	handler := interfacehandlers.NewHandler(log, service)

	playersGroup := v1.Group("/teams/:team_id/players")
	playersGroup.Use(authMiddleware)
	playersGroup.POST("", handler.Create)
	playersGroup.GET("", handler.List)
	playersGroup.GET("/:player_id", handler.GetByID)
	playersGroup.PUT("/:player_id", handler.Update)
	playersGroup.DELETE("/:player_id", handler.Delete)

	return nil
}
