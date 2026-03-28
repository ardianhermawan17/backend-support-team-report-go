package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	authdtos "backend-sport-team-report-go/internal/modules/auth/application/dtos"
	"backend-sport-team-report-go/internal/modules/teams/application"
	teamdomain "backend-sport-team-report-go/internal/modules/teams/domain"
	"backend-sport-team-report-go/internal/modules/teams/interfaces/http/requests"
	"backend-sport-team-report-go/internal/modules/teams/interfaces/http/responses"
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

	var request requests.UpsertTeamRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "invalid request body"})
		return
	}

	team, err := h.service.Create(c.Request.Context(), application.CreateTeamInput{
		CompanyID:             account.CompanyID,
		ActorUserID:           account.UserID,
		Name:                  request.Name,
		LogoImageID:           request.LogoImageID,
		FoundedYear:           request.FoundedYear,
		HomebaseAddress:       request.HomebaseAddress,
		CityOfHomebaseAddress: request.CityOfHomebaseAddress,
	})
	if err != nil {
		h.handleError(c, err, "teams create failed")
		return
	}

	c.JSON(http.StatusCreated, responses.NewTeamResponse(team))
}

func (h *Handler) List(c *gin.Context) {
	account, ok := authenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	teams, err := h.service.List(c.Request.Context(), account.CompanyID)
	if err != nil {
		h.log.InfoContext(c.Request.Context(), "teams list failed", "path", c.FullPath(), "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "unable to list teams"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": responses.NewTeamListResponse(teams)})
}

func (h *Handler) GetByID(c *gin.Context) {
	account, ok := authenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	teamID, ok := parseTeamID(c)
	if !ok {
		return
	}

	team, err := h.service.Get(c.Request.Context(), account.CompanyID, teamID)
	if err != nil {
		h.handleError(c, err, "teams get failed")
		return
	}

	c.JSON(http.StatusOK, responses.NewTeamResponse(team))
}

func (h *Handler) Update(c *gin.Context) {
	account, ok := authenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	teamID, ok := parseTeamID(c)
	if !ok {
		return
	}

	var request requests.UpsertTeamRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "invalid request body"})
		return
	}

	updatedTeam, err := h.service.Update(c.Request.Context(), application.UpdateTeamInput{
		TeamID:                teamID,
		CompanyID:             account.CompanyID,
		ActorUserID:           account.UserID,
		Name:                  request.Name,
		LogoImageID:           request.LogoImageID,
		FoundedYear:           request.FoundedYear,
		HomebaseAddress:       request.HomebaseAddress,
		CityOfHomebaseAddress: request.CityOfHomebaseAddress,
	})
	if err != nil {
		h.handleError(c, err, "teams update failed")
		return
	}

	c.JSON(http.StatusOK, responses.NewTeamResponse(updatedTeam))
}

func (h *Handler) Delete(c *gin.Context) {
	account, ok := authenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	teamID, ok := parseTeamID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), account.CompanyID, teamID, account.UserID); err != nil {
		h.handleError(c, err, "teams delete failed")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) handleError(c *gin.Context, err error, logMessage string) {
	switch {
	case errors.Is(err, teamdomain.ErrInvalidTeamInput):
		message := strings.TrimPrefix(err.Error(), teamdomain.ErrInvalidTeamInput.Error()+": ")
		if message == err.Error() {
			message = "invalid team payload"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": message})
		return
	case errors.Is(err, teamdomain.ErrTeamNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "team_not_found", "message": "team not found in your company"})
		return
	case errors.Is(err, teamdomain.ErrTeamAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "team_name_conflict", "message": "team name already exists in your company"})
		return
	case errors.Is(err, teamdomain.ErrTeamLogoAlreadyInUse):
		c.JSON(http.StatusConflict, gin.H{"error": "team_logo_conflict", "message": "logo image is already in use by another team"})
		return
	default:
		h.log.InfoContext(c.Request.Context(), logMessage, "path", c.FullPath(), "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "unable to process team request"})
	}
}

func parseTeamID(c *gin.Context) (int64, bool) {
	teamID, err := strconv.ParseInt(c.Param("team_id"), 10, 64)
	if err != nil || teamID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "team_id must be a positive bigint"})
		return 0, false
	}

	return teamID, true
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
