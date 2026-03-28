package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	authdtos "backend-sport-team-report-go/internal/modules/auth/application/dtos"
	"backend-sport-team-report-go/internal/modules/schedule/application"
	scheduledomain "backend-sport-team-report-go/internal/modules/schedule/domain"
	"backend-sport-team-report-go/internal/modules/schedule/interfaces/http/requests"
	"backend-sport-team-report-go/internal/modules/schedule/interfaces/http/responses"
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

	var request requests.UpsertScheduleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "invalid request body"})
		return
	}

	matchDate, matchTime, ok := parseScheduleDateTime(c, request.MatchDate, request.MatchTime)
	if !ok {
		return
	}

	schedule, err := h.service.Create(c.Request.Context(), application.CreateScheduleInput{
		CompanyID:   account.CompanyID,
		ActorUserID: account.UserID,
		MatchDate:   matchDate,
		MatchTime:   matchTime,
		HomeTeamID:  request.HomeTeamID,
		GuestTeamID: request.GuestTeamID,
	})
	if err != nil {
		h.handleError(c, err, "schedules create failed")
		return
	}

	c.JSON(http.StatusCreated, responses.NewScheduleResponse(schedule))
}

func (h *Handler) List(c *gin.Context) {
	account, ok := authenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	schedules, err := h.service.List(c.Request.Context(), account.CompanyID)
	if err != nil {
		h.log.InfoContext(c.Request.Context(), "schedules list failed", "path", c.FullPath(), "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "unable to list schedules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": responses.NewScheduleListResponse(schedules)})
}

func (h *Handler) GetByID(c *gin.Context) {
	account, ok := authenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	scheduleID, ok := parseScheduleID(c)
	if !ok {
		return
	}

	schedule, err := h.service.Get(c.Request.Context(), account.CompanyID, scheduleID)
	if err != nil {
		h.handleError(c, err, "schedules get failed")
		return
	}

	c.JSON(http.StatusOK, responses.NewScheduleResponse(schedule))
}

func (h *Handler) Update(c *gin.Context) {
	account, ok := authenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	scheduleID, ok := parseScheduleID(c)
	if !ok {
		return
	}

	var request requests.UpsertScheduleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "invalid request body"})
		return
	}

	matchDate, matchTime, ok := parseScheduleDateTime(c, request.MatchDate, request.MatchTime)
	if !ok {
		return
	}

	schedule, err := h.service.Update(c.Request.Context(), application.UpdateScheduleInput{
		ScheduleID:  scheduleID,
		CompanyID:   account.CompanyID,
		ActorUserID: account.UserID,
		MatchDate:   matchDate,
		MatchTime:   matchTime,
		HomeTeamID:  request.HomeTeamID,
		GuestTeamID: request.GuestTeamID,
	})
	if err != nil {
		h.handleError(c, err, "schedules update failed")
		return
	}

	c.JSON(http.StatusOK, responses.NewScheduleResponse(schedule))
}

func (h *Handler) Delete(c *gin.Context) {
	account, ok := authenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	scheduleID, ok := parseScheduleID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), account.CompanyID, scheduleID, account.UserID); err != nil {
		h.handleError(c, err, "schedules delete failed")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) handleError(c *gin.Context, err error, logMessage string) {
	switch {
	case errors.Is(err, scheduledomain.ErrInvalidScheduleInput):
		message := strings.TrimPrefix(err.Error(), scheduledomain.ErrInvalidScheduleInput.Error()+": ")
		if message == err.Error() {
			message = "invalid schedule payload"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": message})
		return
	case errors.Is(err, scheduledomain.ErrScheduleTeamNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "team_not_found", "message": "home_team_id or guest_team_id was not found in your company"})
		return
	case errors.Is(err, scheduledomain.ErrScheduleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "schedule_not_found", "message": "schedule not found in your company"})
		return
	case errors.Is(err, scheduledomain.ErrScheduleAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "schedule_conflict", "message": "schedule already exists for this match date, time, and team pairing"})
		return
	default:
		h.log.InfoContext(c.Request.Context(), logMessage, "path", c.FullPath(), "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "unable to process schedule request"})
	}
}

func parseScheduleID(c *gin.Context) (int64, bool) {
	scheduleID, err := strconv.ParseInt(c.Param("schedule_id"), 10, 64)
	if err != nil || scheduleID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "schedule_id must be a positive bigint"})
		return 0, false
	}

	return scheduleID, true
}

func parseScheduleDateTime(c *gin.Context, matchDateInput, matchTimeInput string) (time.Time, time.Time, bool) {
	matchDate, err := time.Parse("2006-01-02", strings.TrimSpace(matchDateInput))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "match_date must be in YYYY-MM-DD format"})
		return time.Time{}, time.Time{}, false
	}

	matchTime, err := time.Parse("15:04:05", strings.TrimSpace(matchTimeInput))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "match_time must be in HH:MM:SS format"})
		return time.Time{}, time.Time{}, false
	}

	return matchDate, matchTime, true
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
