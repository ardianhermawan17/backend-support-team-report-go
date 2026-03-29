package http

import (
	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/modules/schedule/application"
	schedulesid "backend-sport-team-report-go/internal/modules/schedule/infrastructure/id"
	schedulespersistence "backend-sport-team-report-go/internal/modules/schedule/infrastructure/persistence"
	interfacehandlers "backend-sport-team-report-go/internal/modules/schedule/interfaces/http/handlers"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(v1 *gin.RouterGroup, db *postgres.Connection, log *logger.Logger, authMiddleware gin.HandlerFunc, security config.SecurityConfig, writeRateLimitMiddleware gin.HandlerFunc) error {
	repository := schedulespersistence.NewScheduleRepository(db)
	idGenerator, err := schedulesid.NewSnowflakeGenerator(3)
	if err != nil {
		return err
	}
	service := application.NewService(repository, idGenerator)
	handler := interfacehandlers.NewHandler(log, service, security.MaxJSONBodyBytes)

	schedulesGroup := v1.Group("/schedules")
	schedulesGroup.Use(authMiddleware)
	schedulesGroup.POST("", writeRateLimitMiddleware, handler.Create)
	schedulesGroup.GET("", handler.List)
	schedulesGroup.GET("/:schedule_id", handler.GetByID)
	schedulesGroup.PUT("/:schedule_id", writeRateLimitMiddleware, handler.Update)
	schedulesGroup.DELETE("/:schedule_id", writeRateLimitMiddleware, handler.Delete)

	return nil
}
