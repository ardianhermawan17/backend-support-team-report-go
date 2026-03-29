package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"backend-sport-team-report-go/internal/modules/player/application/ports"
	playerdomain "backend-sport-team-report-go/internal/modules/player/domain"
	"backend-sport-team-report-go/internal/modules/player/domain/entities"
	"backend-sport-team-report-go/internal/modules/player/domain/repositories"
	"backend-sport-team-report-go/internal/shared/paginator"
)

var validPositions = map[string]struct{}{
	"striker":    {},
	"midfielder": {},
	"defender":   {},
	"goalkeeper": {},
}

type Service struct {
	repository repositories.PlayerRepository
	idGen      ports.IDGenerator
}

type CreatePlayerInput struct {
	CompanyID      int64
	TeamID         int64
	ActorUserID    int64
	Name           string
	Height         float64
	Weight         float64
	Position       string
	PlayerNumber   int
	ProfileImageID *int64
}

type UpdatePlayerInput struct {
	CompanyID      int64
	TeamID         int64
	PlayerID       int64
	ActorUserID    int64
	Name           string
	Height         float64
	Weight         float64
	Position       string
	PlayerNumber   int
	ProfileImageID *int64
}

func NewService(repository repositories.PlayerRepository, idGen ports.IDGenerator) Service {
	return Service{repository: repository, idGen: idGen}
}

func (s Service) Create(ctx context.Context, input CreatePlayerInput) (entities.Player, error) {
	name, position, err := validateInput(input.Name, input.Height, input.Weight, input.Position, input.PlayerNumber, input.ProfileImageID)
	if err != nil {
		return entities.Player{}, err
	}

	id, err := s.idGen.NewID()
	if err != nil {
		return entities.Player{}, fmt.Errorf("generate player id: %w", err)
	}

	player := entities.Player{
		ID:             id,
		TeamID:         input.TeamID,
		Name:           name,
		Height:         input.Height,
		Weight:         input.Weight,
		Position:       position,
		PlayerNumber:   input.PlayerNumber,
		ProfileImageID: input.ProfileImageID,
	}

	if err := s.repository.Create(ctx, input.CompanyID, player, input.ActorUserID); err != nil {
		return entities.Player{}, err
	}

	created, err := s.repository.FindByID(ctx, input.CompanyID, input.TeamID, player.ID)
	if err != nil {
		return entities.Player{}, err
	}

	return created, nil
}

func (s Service) List(ctx context.Context, companyID, teamID int64, params paginator.Params) (paginator.Result[entities.Player], error) {
	return s.repository.ListByTeam(ctx, companyID, teamID, params)
}

func (s Service) Get(ctx context.Context, companyID, teamID, playerID int64) (entities.Player, error) {
	return s.repository.FindByID(ctx, companyID, teamID, playerID)
}

func (s Service) Update(ctx context.Context, input UpdatePlayerInput) (entities.Player, error) {
	existing, err := s.repository.FindByID(ctx, input.CompanyID, input.TeamID, input.PlayerID)
	if err != nil {
		return entities.Player{}, err
	}

	name, position, err := validateInput(input.Name, input.Height, input.Weight, input.Position, input.PlayerNumber, input.ProfileImageID)
	if err != nil {
		return entities.Player{}, err
	}

	existing.Name = name
	existing.Height = input.Height
	existing.Weight = input.Weight
	existing.Position = position
	existing.PlayerNumber = input.PlayerNumber
	existing.ProfileImageID = input.ProfileImageID

	if err := s.repository.Update(ctx, input.CompanyID, existing, input.ActorUserID); err != nil {
		return entities.Player{}, err
	}

	updated, err := s.repository.FindByID(ctx, input.CompanyID, input.TeamID, existing.ID)
	if err != nil {
		return entities.Player{}, err
	}

	return updated, nil
}

func (s Service) Delete(ctx context.Context, companyID, teamID, playerID, actorUserID int64) error {
	deleted, err := s.repository.SoftDelete(ctx, companyID, teamID, playerID, actorUserID)
	if err != nil {
		return err
	}
	if !deleted {
		return playerdomain.ErrPlayerNotFound
	}

	return nil
}

func IsConflictError(err error) bool {
	return errors.Is(err, playerdomain.ErrPlayerNumberAlreadyInUse) || errors.Is(err, playerdomain.ErrPlayerProfileInUse)
}

func validateInput(name string, height, weight float64, position string, playerNumber int, profileImageID *int64) (string, string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "", "", fmt.Errorf("%w: name is required", playerdomain.ErrInvalidPlayerInput)
	}

	if height <= 0 {
		return "", "", fmt.Errorf("%w: height must be greater than 0", playerdomain.ErrInvalidPlayerInput)
	}
	if weight <= 0 {
		return "", "", fmt.Errorf("%w: weight must be greater than 0", playerdomain.ErrInvalidPlayerInput)
	}
	if playerNumber <= 0 || playerNumber > 99 {
		return "", "", fmt.Errorf("%w: player_number must be between 1 and 99", playerdomain.ErrInvalidPlayerInput)
	}
	if profileImageID != nil && *profileImageID <= 0 {
		return "", "", fmt.Errorf("%w: profile_image_id must be a positive bigint", playerdomain.ErrInvalidPlayerInput)
	}

	normalizedPosition := strings.ToLower(strings.TrimSpace(position))
	if _, ok := validPositions[normalizedPosition]; !ok {
		return "", "", fmt.Errorf("%w: position must be one of striker|midfielder|defender|goalkeeper", playerdomain.ErrInvalidPlayerInput)
	}

	return trimmedName, normalizedPosition, nil
}
