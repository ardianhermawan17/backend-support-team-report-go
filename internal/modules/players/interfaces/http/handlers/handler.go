package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	authdtos "backend-sport-team-report-go/internal/modules/auth/application/dtos"
	"backend-sport-team-report-go/internal/modules/players/application"
	playerdomain "backend-sport-team-report-go/internal/modules/players/domain"
	"backend-sport-team-report-go/internal/modules/players/interfaces/http/requests"
	"backend-sport-team-report-go/internal/modules/players/interfaces/http/responses"
	"backend-sport-team-report-go/internal/shared/logger"
)

const authenticatedAccountContextKey = "auth.account"

type Handler struct {
	log     *logger.Logger
	service application.Service
}

func NewHandler(log *logger.Logger, service application.Service) *Handler {
	return &Handler{log: log, service: service}
}

func (h *Handler) Create(c *gin.Context) {
	account, ok := authenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	teamID, ok := parsePositiveBigIntPath(c, "team_id")
	if !ok {
		return
	}

	var request requests.UpsertPlayerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "invalid request body"})
		return
	}

	player, err := h.service.Create(c.Request.Context(), application.CreatePlayerInput{
		CompanyID:      account.CompanyID,
		TeamID:         teamID,
		ActorUserID:    account.UserID,
		Name:           request.Name,
		Height:         request.Height,
		Weight:         request.Weight,
		Position:       request.Position,
		PlayerNumber:   request.PlayerNumber,
		ProfileImageID: request.ProfileImageID,
	})
	if err != nil {
		h.handleError(c, err, "players create failed")
		return
	}

	c.JSON(http.StatusCreated, responses.NewPlayerResponse(player))
}

func (h *Handler) List(c *gin.Context) {
	account, ok := authenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	teamID, ok := parsePositiveBigIntPath(c, "team_id")
	if !ok {
		return
	}

	players, err := h.service.List(c.Request.Context(), account.CompanyID, teamID)
	if err != nil {
		h.handleError(c, err, "players list failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": responses.NewPlayerListResponse(players)})
}

func (h *Handler) GetByID(c *gin.Context) {
	account, ok := authenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	teamID, ok := parsePositiveBigIntPath(c, "team_id")
	if !ok {
		return
	}

	playerID, ok := parsePositiveBigIntPath(c, "player_id")
	if !ok {
		return
	}

	player, err := h.service.Get(c.Request.Context(), account.CompanyID, teamID, playerID)
	if err != nil {
		h.handleError(c, err, "players get failed")
		return
	}

	c.JSON(http.StatusOK, responses.NewPlayerResponse(player))
}

func (h *Handler) Update(c *gin.Context) {
	account, ok := authenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	teamID, ok := parsePositiveBigIntPath(c, "team_id")
	if !ok {
		return
	}

	playerID, ok := parsePositiveBigIntPath(c, "player_id")
	if !ok {
		return
	}

	var request requests.UpsertPlayerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "invalid request body"})
		return
	}

	player, err := h.service.Update(c.Request.Context(), application.UpdatePlayerInput{
		CompanyID:      account.CompanyID,
		TeamID:         teamID,
		PlayerID:       playerID,
		ActorUserID:    account.UserID,
		Name:           request.Name,
		Height:         request.Height,
		Weight:         request.Weight,
		Position:       request.Position,
		PlayerNumber:   request.PlayerNumber,
		ProfileImageID: request.ProfileImageID,
	})
	if err != nil {
		h.handleError(c, err, "players update failed")
		return
	}

	c.JSON(http.StatusOK, responses.NewPlayerResponse(player))
}

func (h *Handler) Delete(c *gin.Context) {
	account, ok := authenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	teamID, ok := parsePositiveBigIntPath(c, "team_id")
	if !ok {
		return
	}

	playerID, ok := parsePositiveBigIntPath(c, "player_id")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), account.CompanyID, teamID, playerID, account.UserID); err != nil {
		h.handleError(c, err, "players delete failed")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) handleError(c *gin.Context, err error, logMessage string) {
	switch {
	case errors.Is(err, playerdomain.ErrInvalidPlayerInput):
		message := strings.TrimPrefix(err.Error(), playerdomain.ErrInvalidPlayerInput.Error()+": ")
		if message == err.Error() {
			message = "invalid player payload"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": message})
		return
	case errors.Is(err, playerdomain.ErrTeamNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "team_not_found", "message": "team not found in your company"})
		return
	case errors.Is(err, playerdomain.ErrPlayerNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "player_not_found", "message": "player not found in your team"})
		return
	case errors.Is(err, playerdomain.ErrPlayerNumberAlreadyInUse):
		c.JSON(http.StatusConflict, gin.H{"error": "player_number_conflict", "message": "player number already exists in this team"})
		return
	case errors.Is(err, playerdomain.ErrPlayerProfileInUse):
		c.JSON(http.StatusConflict, gin.H{"error": "profile_image_conflict", "message": "profile image is already in use by another player"})
		return
	default:
		h.log.InfoContext(c.Request.Context(), logMessage, "path", c.FullPath(), "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "unable to process player request"})
	}
}

func parsePositiveBigIntPath(c *gin.Context, key string) (int64, bool) {
	value, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil || value <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": key + " must be a positive bigint"})
		return 0, false
	}

	return value, true
}

func authenticatedAccount(c *gin.Context) (authdtos.AuthenticatedAccount, bool) {
	value, ok := c.Get(authenticatedAccountContextKey)
	if !ok {
		return authdtos.AuthenticatedAccount{}, false
	}

	account, ok := value.(authdtos.AuthenticatedAccount)
	if !ok {
		return authdtos.AuthenticatedAccount{}, false
	}

	return account, true
}
