package http

import (
	"backend-sport-team-report-go/internal/modules/team/application"
	teamsid "backend-sport-team-report-go/internal/modules/team/infrastructure/id"
	teamspersistence "backend-sport-team-report-go/internal/modules/team/infrastructure/persistence"
	interfacehandlers "backend-sport-team-report-go/internal/modules/team/interfaces/http/handlers"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(v1 *gin.RouterGroup, db *postgres.Connection, log *logger.Logger, authMiddleware gin.HandlerFunc) error {
	repository := teamspersistence.NewTeamRepository(db)
	idGenerator, err := teamsid.NewSnowflakeGenerator(1)
	if err != nil {
		return err
	}
	service := application.NewService(repository, idGenerator)
	handler := interfacehandlers.NewHandler(log, service)

	teamsGroup := v1.Group("/teams")
	teamsGroup.Use(authMiddleware)
	teamsGroup.POST("", handler.Create)
	teamsGroup.GET("", handler.List)
	teamsGroup.GET("/:team_id", handler.GetByID)
	teamsGroup.PUT("/:team_id", handler.Update)
	teamsGroup.DELETE("/:team_id", handler.Delete)

	return nil
}
