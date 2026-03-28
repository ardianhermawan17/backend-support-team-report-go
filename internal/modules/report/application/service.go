package application

import (
	"context"
	"errors"
	"fmt"

	"backend-sport-team-report-go/internal/modules/report/application/ports"
	reportdomain "backend-sport-team-report-go/internal/modules/report/domain"
	"backend-sport-team-report-go/internal/modules/report/domain/entities"
	"backend-sport-team-report-go/internal/modules/report/domain/repositories"
)

const (
	StatusHomeTeamWin  = "home_team_win"
	StatusGuestTeamWin = "guest_team_win"
	StatusDraw         = "draw"
)

type Service struct {
	repository repositories.ReportRepository
	idGen      ports.IDGenerator
}

type CreateReportInput struct {
	CompanyID   int64
	ActorUserID int64

	MatchScheduleID         int64
	FinalScoreHome          int
	FinalScoreGuest         int
	MostScoringGoalPlayerID *int64
}

type UpdateReportInput struct {
	ReportID    int64
	CompanyID   int64
	ActorUserID int64

	MatchScheduleID         int64
	FinalScoreHome          int
	FinalScoreGuest         int
	MostScoringGoalPlayerID *int64
}

func NewService(repository repositories.ReportRepository, idGen ports.IDGenerator) Service {
	return Service{repository: repository, idGen: idGen}
}

func (s Service) Create(ctx context.Context, input CreateReportInput) (entities.Report, error) {
	if err := validateInput(input.MatchScheduleID, input.FinalScoreHome, input.FinalScoreGuest, input.MostScoringGoalPlayerID); err != nil {
		return entities.Report{}, err
	}

	id, err := s.idGen.NewID()
	if err != nil {
		return entities.Report{}, fmt.Errorf("generate report id: %w", err)
	}

	report := entities.Report{
		ID:                      id,
		CompanyID:               input.CompanyID,
		MatchScheduleID:         input.MatchScheduleID,
		FinalScoreHome:          input.FinalScoreHome,
		FinalScoreGuest:         input.FinalScoreGuest,
		StatusMatch:             calculateStatus(input.FinalScoreHome, input.FinalScoreGuest),
		MostScoringGoalPlayerID: input.MostScoringGoalPlayerID,
	}

	if err := s.repository.Create(ctx, report, input.ActorUserID); err != nil {
		return entities.Report{}, err
	}

	created, err := s.repository.FindByIDAndCompany(ctx, report.ID, report.CompanyID)
	if err != nil {
		return entities.Report{}, err
	}

	return created, nil
}

func (s Service) List(ctx context.Context, companyID int64) ([]entities.Report, error) {
	return s.repository.ListByCompany(ctx, companyID)
}

func (s Service) Get(ctx context.Context, companyID, reportID int64) (entities.Report, error) {
	return s.repository.FindByIDAndCompany(ctx, reportID, companyID)
}

func (s Service) Update(ctx context.Context, input UpdateReportInput) (entities.Report, error) {
	existing, err := s.repository.FindByIDAndCompany(ctx, input.ReportID, input.CompanyID)
	if err != nil {
		return entities.Report{}, err
	}

	if err := validateInput(input.MatchScheduleID, input.FinalScoreHome, input.FinalScoreGuest, input.MostScoringGoalPlayerID); err != nil {
		return entities.Report{}, err
	}

	existing.MatchScheduleID = input.MatchScheduleID
	existing.FinalScoreHome = input.FinalScoreHome
	existing.FinalScoreGuest = input.FinalScoreGuest
	existing.StatusMatch = calculateStatus(input.FinalScoreHome, input.FinalScoreGuest)
	existing.MostScoringGoalPlayerID = input.MostScoringGoalPlayerID

	if err := s.repository.Update(ctx, existing, input.ActorUserID); err != nil {
		return entities.Report{}, err
	}

	updated, err := s.repository.FindByIDAndCompany(ctx, existing.ID, existing.CompanyID)
	if err != nil {
		return entities.Report{}, err
	}

	return updated, nil
}

func (s Service) Delete(ctx context.Context, companyID, reportID, actorUserID int64) error {
	deleted, err := s.repository.SoftDelete(ctx, reportID, companyID, actorUserID)
	if err != nil {
		return err
	}
	if !deleted {
		return reportdomain.ErrReportNotFound
	}

	return nil
}

func IsConflictError(err error) bool {
	return errors.Is(err, reportdomain.ErrReportAlreadyExists)
}

func validateInput(matchScheduleID int64, finalScoreHome, finalScoreGuest int, mostScoringGoalPlayerID *int64) error {
	if matchScheduleID <= 0 {
		return fmt.Errorf("%w: match_schedule_id must be a positive bigint", reportdomain.ErrInvalidReportInput)
	}
	if finalScoreHome < 0 {
		return fmt.Errorf("%w: final_score_home must be greater than or equal to 0", reportdomain.ErrInvalidReportInput)
	}
	if finalScoreGuest < 0 {
		return fmt.Errorf("%w: final_score_guest must be greater than or equal to 0", reportdomain.ErrInvalidReportInput)
	}
	if mostScoringGoalPlayerID != nil && *mostScoringGoalPlayerID <= 0 {
		return fmt.Errorf("%w: most_scoring_goal_player_id must be a positive bigint", reportdomain.ErrInvalidReportInput)
	}

	return nil
}

func calculateStatus(finalScoreHome, finalScoreGuest int) string {
	switch {
	case finalScoreHome > finalScoreGuest:
		return StatusHomeTeamWin
	case finalScoreGuest > finalScoreHome:
		return StatusGuestTeamWin
	default:
		return StatusDraw
	}
}
