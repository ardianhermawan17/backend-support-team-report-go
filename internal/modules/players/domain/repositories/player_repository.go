package repositories

import (
	"context"

	"backend-sport-team-report-go/internal/modules/players/domain/entities"
)

type PlayerRepository interface {
	Create(ctx context.Context, companyID int64, player entities.Player, actorUserID int64) error
	ListByTeam(ctx context.Context, companyID, teamID int64) ([]entities.Player, error)
	FindByID(ctx context.Context, companyID, teamID, playerID int64) (entities.Player, error)
	Update(ctx context.Context, companyID int64, player entities.Player, actorUserID int64) error
	SoftDelete(ctx context.Context, companyID, teamID, playerID, actorUserID int64) (bool, error)
}
