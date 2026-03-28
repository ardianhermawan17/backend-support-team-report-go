package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"backend-sport-team-report-go/internal/modules/auth/application/dtos"
	authapplication "backend-sport-team-report-go/internal/modules/auth/application/handlers"
	"backend-sport-team-report-go/internal/modules/auth/application/ports"
	authdomain "backend-sport-team-report-go/internal/modules/auth/domain"
	"backend-sport-team-report-go/internal/modules/auth/interfaces/http/requests"
	"backend-sport-team-report-go/internal/modules/auth/interfaces/http/responses"
	"backend-sport-team-report-go/internal/shared/logger"
)

const authenticatedAccountContextKey = "auth.account"

type Handler struct {
	log            *logger.Logger
	login          authapplication.LoginHandler
	currentAccount authapplication.CurrentAccountHandler
	tokens         ports.TokenService
}

func NewHandler(log *logger.Logger, login authapplication.LoginHandler, currentAccount authapplication.CurrentAccountHandler, tokens ports.TokenService) *Handler {
	return &Handler{
		log:            log,
		login:          login,
		currentAccount: currentAccount,
		tokens:         tokens,
	}
}

func (h *Handler) Login(c *gin.Context) {
	var request requests.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "username and password are required"})
		return
	}

	request.Username = strings.TrimSpace(request.Username)
	if request.Username == "" || request.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "username and password are required"})
		return
	}

	result, err := h.login.Handle(c.Request.Context(), request.Username, request.Password)
	if err != nil {
		if errors.Is(err, authdomain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials", "message": "username or password is incorrect"})
			return
		}

		h.log.InfoContext(c.Request.Context(), "auth login failed", "path", c.FullPath(), "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "unable to complete login"})
		return
	}

	c.JSON(http.StatusOK, responses.NewLoginResponse(result))
}

func (h *Handler) Me(c *gin.Context) {
	account, ok := c.Get(authenticatedAccountContextKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	authenticatedAccount, ok := account.(dtos.AuthenticatedAccount)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	c.JSON(http.StatusOK, responses.NewCurrentAccountResponse(authenticatedAccount))
}

func (h *Handler) RequireAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "bearer token is required"})
			return
		}

		identity, err := h.tokens.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "bearer token is invalid"})
			return
		}

		account, err := h.currentAccount.Handle(c.Request.Context(), identity)
		if err != nil {
			if errors.Is(err, authdomain.ErrUnauthorized) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authenticated account is no longer active"})
				return
			}

			h.log.InfoContext(c.Request.Context(), "auth current account lookup failed", "path", c.FullPath(), "error", err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "unable to resolve authenticated account"})
			return
		}

		c.Set(authenticatedAccountContextKey, account)
		c.Next()
	}
}

func bearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}

	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}

	return strings.TrimSpace(parts[1]), true
}
