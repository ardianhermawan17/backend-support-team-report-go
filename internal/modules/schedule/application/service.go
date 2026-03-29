package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend-sport-team-report-go/internal/modules/schedule/application/ports"
	scheduledomain "backend-sport-team-report-go/internal/modules/schedule/domain"
	"backend-sport-team-report-go/internal/modules/schedule/domain/entities"
	"backend-sport-team-report-go/internal/modules/schedule/domain/repositories"
	"backend-sport-team-report-go/internal/shared/paginator"
)

type Service struct {
	repository repositories.ScheduleRepository
	idGen      ports.IDGenerator
	writeGate  *scheduleWriteGate
}

type CreateScheduleInput struct {
	CompanyID   int64
	ActorUserID int64
	MatchDate   time.Time
	MatchTime   time.Time
	HomeTeamID  int64
	GuestTeamID int64
}

type UpdateScheduleInput struct {
	ScheduleID  int64
	CompanyID   int64
	ActorUserID int64
	MatchDate   time.Time
	MatchTime   time.Time
	HomeTeamID  int64
	GuestTeamID int64
}

func NewService(repository repositories.ScheduleRepository, idGen ports.IDGenerator) Service {
	return Service{repository: repository, idGen: idGen, writeGate: newScheduleWriteGate()}
}

func (s Service) Create(ctx context.Context, input CreateScheduleInput) (entities.Schedule, error) {
	if err := validateInput(input.MatchDate, input.MatchTime, input.HomeTeamID, input.GuestTeamID); err != nil {
		return entities.Schedule{}, err
	}

	gateKey := scheduleCreateGateKey(input.CompanyID, input.HomeTeamID, input.GuestTeamID, input.MatchDate, input.MatchTime)
	var created entities.Schedule
	err := s.writeGate.WithKey(ctx, gateKey, func() error {
		id, err := s.idGen.NewID()
		if err != nil {
			return fmt.Errorf("generate schedule id: %w", err)
		}

		schedule := entities.Schedule{
			ID:          id,
			CompanyID:   input.CompanyID,
			MatchDate:   input.MatchDate,
			MatchTime:   input.MatchTime,
			HomeTeamID:  input.HomeTeamID,
			GuestTeamID: input.GuestTeamID,
		}

		if err := s.repository.Create(ctx, schedule, input.ActorUserID); err != nil {
			return err
		}

		createdSchedule, err := s.repository.FindByIDAndCompany(ctx, schedule.ID, schedule.CompanyID)
		if err != nil {
			return err
		}

		created = createdSchedule
		return nil
	})
	if err != nil {
		return entities.Schedule{}, err
	}

	return created, nil
}

func (s Service) List(ctx context.Context, companyID int64, params paginator.Params) (paginator.Result[entities.Schedule], error) {
	return s.repository.ListByCompany(ctx, companyID, params)
}

func (s Service) Get(ctx context.Context, companyID, scheduleID int64) (entities.Schedule, error) {
	return s.repository.FindByIDAndCompany(ctx, scheduleID, companyID)
}

func (s Service) Update(ctx context.Context, input UpdateScheduleInput) (entities.Schedule, error) {
	existing, err := s.repository.FindByIDAndCompany(ctx, input.ScheduleID, input.CompanyID)
	if err != nil {
		return entities.Schedule{}, err
	}

	if err := validateInput(input.MatchDate, input.MatchTime, input.HomeTeamID, input.GuestTeamID); err != nil {
		return entities.Schedule{}, err
	}

	existing.MatchDate = input.MatchDate
	existing.MatchTime = input.MatchTime
	existing.HomeTeamID = input.HomeTeamID
	existing.GuestTeamID = input.GuestTeamID

	if err := s.repository.Update(ctx, existing, input.ActorUserID); err != nil {
		return entities.Schedule{}, err
	}

	updated, err := s.repository.FindByIDAndCompany(ctx, existing.ID, existing.CompanyID)
	if err != nil {
		return entities.Schedule{}, err
	}

	return updated, nil
}

func (s Service) Delete(ctx context.Context, companyID, scheduleID, actorUserID int64) error {
	deleted, err := s.repository.SoftDelete(ctx, scheduleID, companyID, actorUserID)
	if err != nil {
		return err
	}
	if !deleted {
		return scheduledomain.ErrScheduleNotFound
	}

	return nil
}

func IsConflictError(err error) bool {
	return errors.Is(err, scheduledomain.ErrScheduleAlreadyExists) || errors.Is(err, scheduledomain.ErrScheduleConcurrentModification)
}

func validateInput(matchDate, matchTime time.Time, homeTeamID, guestTeamID int64) error {
	if matchDate.IsZero() {
		return fmt.Errorf("%w: match_date is required", scheduledomain.ErrInvalidScheduleInput)
	}
	if matchTime.IsZero() {
		return fmt.Errorf("%w: match_time is required", scheduledomain.ErrInvalidScheduleInput)
	}
	if homeTeamID <= 0 {
		return fmt.Errorf("%w: home_team_id must be a positive bigint", scheduledomain.ErrInvalidScheduleInput)
	}
	if guestTeamID <= 0 {
		return fmt.Errorf("%w: guest_team_id must be a positive bigint", scheduledomain.ErrInvalidScheduleInput)
	}
	if homeTeamID == guestTeamID {
		return fmt.Errorf("%w: home_team_id and guest_team_id must be different", scheduledomain.ErrInvalidScheduleInput)
	}

	return nil
}
