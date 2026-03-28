package repositories

import (
	"context"

	"backend-sport-team-report-go/internal/modules/teams/domain/entities"
)

type TeamRepository interface {
	Create(ctx context.Context, team entities.Team, actorUserID int64) error
	ListByCompany(ctx context.Context, companyID int64) ([]entities.Team, error)
	FindByIDAndCompany(ctx context.Context, teamID, companyID int64) (entities.Team, error)
	Update(ctx context.Context, team entities.Team, actorUserID int64) error
	SoftDelete(ctx context.Context, teamID, companyID, actorUserID int64) (bool, error)
}
