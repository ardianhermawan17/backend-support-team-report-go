package handlers

import (
	"errors"
	"net/http"
	"strings"

	authapplication "backend-sport-team-report-go/internal/modules/auth/application/handlers"
	"backend-sport-team-report-go/internal/modules/auth/application/ports"
	authdomain "backend-sport-team-report-go/internal/modules/auth/domain"
	"backend-sport-team-report-go/internal/modules/auth/interfaces/http/requests"
	"backend-sport-team-report-go/internal/modules/auth/interfaces/http/responses"
	"backend-sport-team-report-go/internal/shared/httpjson"
	"backend-sport-team-report-go/internal/shared/logger"
	sharedmiddleware "backend-sport-team-report-go/internal/shared/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	log            *logger.Logger
	login          authapplication.LoginHandler
	currentAccount authapplication.CurrentAccountHandler
	tokens         ports.TokenService
	maxBodyBytes   int64
}

func NewHandler(log *logger.Logger, login authapplication.LoginHandler, currentAccount authapplication.CurrentAccountHandler, tokens ports.TokenService, maxBodyBytes int64) *Handler {
	return &Handler{
		log:            log,
		login:          login,
		currentAccount: currentAccount,
		tokens:         tokens,
		maxBodyBytes:   maxBodyBytes,
	}
}

func (h *Handler) Login(c *gin.Context) {
	var request requests.LoginRequest
	if err := httpjson.Bind(c, &request, h.maxBodyBytes); err != nil {
		if !httpjson.WriteError(c, err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "username and password are required"})
		}
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
			h.log.InfoContext(c.Request.Context(), "auth login rejected", "path", c.FullPath(), "remote_ip", sharedmiddleware.RemoteIP(c))
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
	authenticatedAccount, ok := sharedmiddleware.AuthenticatedAccount(c)
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
			h.log.InfoContext(c.Request.Context(), "auth missing bearer token", "path", c.FullPath(), "remote_ip", sharedmiddleware.RemoteIP(c))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "bearer token is required"})
			return
		}

		identity, err := h.tokens.Parse(token)
		if err != nil {
			h.log.InfoContext(c.Request.Context(), "auth invalid bearer token", "path", c.FullPath(), "remote_ip", sharedmiddleware.RemoteIP(c))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "bearer token is invalid"})
			return
		}

		account, err := h.currentAccount.Handle(c.Request.Context(), identity)
		if err != nil {
			if errors.Is(err, authdomain.ErrUnauthorized) {
				h.log.InfoContext(c.Request.Context(), "auth inactive account rejected", "path", c.FullPath(), "remote_ip", sharedmiddleware.RemoteIP(c))
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authenticated account is no longer active"})
				return
			}

			h.log.InfoContext(c.Request.Context(), "auth current account lookup failed", "path", c.FullPath(), "error", err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "unable to resolve authenticated account"})
			return
		}

		sharedmiddleware.SetAuthenticatedAccount(c, account)
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
