package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend-sport-team-report-go/internal/modules/team/application/ports"
	teamdomain "backend-sport-team-report-go/internal/modules/team/domain"
	"backend-sport-team-report-go/internal/modules/team/domain/entities"
	"backend-sport-team-report-go/internal/modules/team/domain/repositories"
	"backend-sport-team-report-go/internal/shared/paginator"
)

type Service struct {
	repository repositories.TeamRepository
	idGen      ports.IDGenerator
}

type CreateTeamInput struct {
	CompanyID             int64
	ActorUserID           int64
	Name                  string
	LogoImageID           *int64
	FoundedYear           int
	HomebaseAddress       string
	CityOfHomebaseAddress string
}

type UpdateTeamInput struct {
	TeamID                int64
	CompanyID             int64
	ActorUserID           int64
	Name                  string
	LogoImageID           *int64
	FoundedYear           int
	HomebaseAddress       string
	CityOfHomebaseAddress string
}

func NewService(repository repositories.TeamRepository, idGen ports.IDGenerator) Service {
	return Service{repository: repository, idGen: idGen}
}

func (s Service) Create(ctx context.Context, input CreateTeamInput) (entities.Team, error) {
	name := strings.TrimSpace(input.Name)
	homebaseAddress := strings.TrimSpace(input.HomebaseAddress)
	city := strings.TrimSpace(input.CityOfHomebaseAddress)

	if err := validateInput(name, input.LogoImageID, input.FoundedYear, homebaseAddress, city); err != nil {
		return entities.Team{}, err
	}

	id, err := s.idGen.NewID()
	if err != nil {
		return entities.Team{}, fmt.Errorf("generate team id: %w", err)
	}

	team := entities.Team{
		ID:                    id,
		CompanyID:             input.CompanyID,
		Name:                  name,
		LogoImageID:           input.LogoImageID,
		FoundedYear:           input.FoundedYear,
		HomebaseAddress:       homebaseAddress,
		CityOfHomebaseAddress: city,
	}

	if err := s.repository.Create(ctx, team, input.ActorUserID); err != nil {
		return entities.Team{}, err
	}

	created, err := s.repository.FindByIDAndCompany(ctx, team.ID, team.CompanyID)
	if err != nil {
		return entities.Team{}, err
	}

	return created, nil
}

func (s Service) List(ctx context.Context, companyID int64, params paginator.Params) (paginator.Result[entities.Team], error) {
	return s.repository.ListByCompany(ctx, companyID, params)
}

func (s Service) Get(ctx context.Context, companyID, teamID int64) (entities.Team, error) {
	return s.repository.FindByIDAndCompany(ctx, teamID, companyID)
}

func (s Service) Update(ctx context.Context, input UpdateTeamInput) (entities.Team, error) {
	existing, err := s.repository.FindByIDAndCompany(ctx, input.TeamID, input.CompanyID)
	if err != nil {
		return entities.Team{}, err
	}

	name := strings.TrimSpace(input.Name)
	homebaseAddress := strings.TrimSpace(input.HomebaseAddress)
	city := strings.TrimSpace(input.CityOfHomebaseAddress)
	if err := validateInput(name, input.LogoImageID, input.FoundedYear, homebaseAddress, city); err != nil {
		return entities.Team{}, err
	}

	existing.Name = name
	existing.LogoImageID = input.LogoImageID
	existing.FoundedYear = input.FoundedYear
	existing.HomebaseAddress = homebaseAddress
	existing.CityOfHomebaseAddress = city

	if err := s.repository.Update(ctx, existing, input.ActorUserID); err != nil {
		return entities.Team{}, err
	}

	updated, err := s.repository.FindByIDAndCompany(ctx, existing.ID, existing.CompanyID)
	if err != nil {
		return entities.Team{}, err
	}

	return updated, nil
}

func (s Service) Delete(ctx context.Context, companyID, teamID, actorUserID int64) error {
	deleted, err := s.repository.SoftDelete(ctx, teamID, companyID, actorUserID)
	if err != nil {
		return err
	}
	if !deleted {
		return teamdomain.ErrTeamNotFound
	}

	return nil
}

func validateInput(name string, logoImageID *int64, foundedYear int, homebaseAddress, city string) error {
	if name == "" {
		return fmt.Errorf("%w: name is required", teamdomain.ErrInvalidTeamInput)
	}
	if foundedYear < 1800 || foundedYear > time.Now().Year()+1 {
		return fmt.Errorf("%w: founded_year is out of range", teamdomain.ErrInvalidTeamInput)
	}
	if homebaseAddress == "" {
		return fmt.Errorf("%w: homebase_address is required", teamdomain.ErrInvalidTeamInput)
	}
	if city == "" {
		return fmt.Errorf("%w: city_of_homebase_address is required", teamdomain.ErrInvalidTeamInput)
	}
	if logoImageID != nil && *logoImageID <= 0 {
		return fmt.Errorf("%w: logo_image_id must be a positive bigint", teamdomain.ErrInvalidTeamInput)
	}

	return nil
}

func IsConflictError(err error) bool {
	return errors.Is(err, teamdomain.ErrTeamAlreadyExists) || errors.Is(err, teamdomain.ErrTeamLogoAlreadyInUse)
}
