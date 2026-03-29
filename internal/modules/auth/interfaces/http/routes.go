package http

import (
	"backend-sport-team-report-go/internal/config"
	applicationhandlers "backend-sport-team-report-go/internal/modules/auth/application/handlers"
	authjwt "backend-sport-team-report-go/internal/modules/auth/infrastructure/jwt"
	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	interfacehandlers "backend-sport-team-report-go/internal/modules/auth/interfaces/http/handlers"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(v1 *gin.RouterGroup, db *postgres.Connection, log *logger.Logger, cfg config.AuthConfig, security config.SecurityConfig, loginRateLimitMiddleware gin.HandlerFunc) gin.HandlerFunc {
	repository := authpersistence.NewAccountRepository(db)
	tokens := authjwt.NewTokenService(cfg)
	loginHandler := applicationhandlers.NewLoginHandler(repository, tokens)
	currentAccountHandler := applicationhandlers.NewCurrentAccountHandler(repository)
	httpHandler := interfacehandlers.NewHandler(log, loginHandler, currentAccountHandler, tokens, security.MaxJSONBodyBytes)

	authGroup := v1.Group("/auth")
	authGroup.POST("/login", loginRateLimitMiddleware, httpHandler.Login)
	authGroup.GET("/me", httpHandler.RequireAuthentication(), httpHandler.Me)

	return httpHandler.RequireAuthentication()
}
