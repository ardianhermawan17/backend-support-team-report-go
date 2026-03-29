package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"backend-sport-team-report-go/internal/modules/report/application"
	reportdomain "backend-sport-team-report-go/internal/modules/report/domain"
	"backend-sport-team-report-go/internal/modules/report/interfaces/http/requests"
	"backend-sport-team-report-go/internal/modules/report/interfaces/http/responses"
	"backend-sport-team-report-go/internal/shared/logger"
	sharedmiddleware "backend-sport-team-report-go/internal/shared/middleware"
	"backend-sport-team-report-go/internal/shared/paginator"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	log     *logger.Logger
	service application.Service
}

func NewHandler(log *logger.Logger, service application.Service) *Handler {
	return &Handler{log: log, service: service}
}

func (h *Handler) Create(c *gin.Context) {
	account, ok := sharedmiddleware.AuthenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	var request requests.UpsertReportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "invalid request body"})
		return
	}

	report, err := h.service.Create(c.Request.Context(), application.CreateReportInput{
		CompanyID:               account.CompanyID,
		ActorUserID:             account.UserID,
		MatchScheduleID:         request.MatchScheduleID,
		FinalScoreHome:          request.FinalScoreHome,
		FinalScoreGuest:         request.FinalScoreGuest,
		MostScoringGoalPlayerID: request.MostScoringGoalPlayerID,
	})
	if err != nil {
		h.handleError(c, err, "reports create failed")
		return
	}

	c.JSON(http.StatusCreated, responses.NewReportResponse(report))
}

func (h *Handler) List(c *gin.Context) {
	account, ok := sharedmiddleware.AuthenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	params, err := paginator.FromRaw(c.Query("page"), c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "page and limit must be positive integers"})
		return
	}

	reportsData, err := h.service.List(c.Request.Context(), account.CompanyID, params)
	if err != nil {
		h.log.InfoContext(c.Request.Context(), "reports list failed", "path", c.FullPath(), "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "unable to list reports"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": responses.NewReportListResponse(reportsData.Items), "meta": reportsData.Meta})
}

func (h *Handler) GetByID(c *gin.Context) {
	account, ok := sharedmiddleware.AuthenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	reportID, ok := parseReportID(c)
	if !ok {
		return
	}

	report, err := h.service.Get(c.Request.Context(), account.CompanyID, reportID)
	if err != nil {
		h.handleError(c, err, "reports get failed")
		return
	}

	c.JSON(http.StatusOK, responses.NewReportResponse(report))
}

func (h *Handler) Update(c *gin.Context) {
	account, ok := sharedmiddleware.AuthenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	reportID, ok := parseReportID(c)
	if !ok {
		return
	}

	var request requests.UpsertReportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "invalid request body"})
		return
	}

	report, err := h.service.Update(c.Request.Context(), application.UpdateReportInput{
		ReportID:                reportID,
		CompanyID:               account.CompanyID,
		ActorUserID:             account.UserID,
		MatchScheduleID:         request.MatchScheduleID,
		FinalScoreHome:          request.FinalScoreHome,
		FinalScoreGuest:         request.FinalScoreGuest,
		MostScoringGoalPlayerID: request.MostScoringGoalPlayerID,
	})
	if err != nil {
		h.handleError(c, err, "reports update failed")
		return
	}

	c.JSON(http.StatusOK, responses.NewReportResponse(report))
}

func (h *Handler) Delete(c *gin.Context) {
	account, ok := sharedmiddleware.AuthenticatedAccount(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication is required"})
		return
	}

	reportID, ok := parseReportID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), account.CompanyID, reportID, account.UserID); err != nil {
		h.handleError(c, err, "reports delete failed")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) handleError(c *gin.Context, err error, logMessage string) {
	switch {
	case errors.Is(err, reportdomain.ErrInvalidReportInput):
		message := strings.TrimPrefix(err.Error(), reportdomain.ErrInvalidReportInput.Error()+": ")
		if message == err.Error() {
			message = "invalid report payload"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": message})
		return
	case errors.Is(err, reportdomain.ErrReportNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "report_not_found", "message": "report not found in your company"})
		return
	case errors.Is(err, reportdomain.ErrReportScheduleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "schedule_not_found", "message": "match_schedule_id was not found in your company"})
		return
	case errors.Is(err, reportdomain.ErrReportTopScorerNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "top_scorer_not_found", "message": "most_scoring_goal_player_id was not found in this match teams"})
		return
	case errors.Is(err, reportdomain.ErrReportAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "report_conflict", "message": "report already exists for this match schedule"})
		return
	default:
		h.log.InfoContext(c.Request.Context(), logMessage, "path", c.FullPath(), "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "unable to process report request"})
	}
}

func parseReportID(c *gin.Context) (int64, bool) {
	reportID, err := strconv.ParseInt(c.Param("report_id"), 10, 64)
	if err != nil || reportID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "report_id must be a positive bigint"})
		return 0, false
	}

	return reportID, true
}
