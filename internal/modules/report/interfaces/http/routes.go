package http

import (
	"backend-sport-team-report-go/internal/modules/report/application"
	reportsid "backend-sport-team-report-go/internal/modules/report/infrastructure/id"
	reportspersistence "backend-sport-team-report-go/internal/modules/report/infrastructure/persistence"
	interfacehandlers "backend-sport-team-report-go/internal/modules/report/interfaces/http/handlers"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(v1 *gin.RouterGroup, db *postgres.Connection, log *logger.Logger, authMiddleware gin.HandlerFunc) error {
	repository := reportspersistence.NewReportRepository(db)
	idGenerator, err := reportsid.NewSnowflakeGenerator(4)
	if err != nil {
		return err
	}
	service := application.NewService(repository, idGenerator)
	handler := interfacehandlers.NewHandler(log, service)

	reportsGroup := v1.Group("/reports")
	reportsGroup.Use(authMiddleware)
	reportsGroup.POST("", handler.Create)
	reportsGroup.GET("", handler.List)
	reportsGroup.GET("/:report_id", handler.GetByID)
	reportsGroup.PUT("/:report_id", handler.Update)
	reportsGroup.DELETE("/:report_id", handler.Delete)

	return nil
}
