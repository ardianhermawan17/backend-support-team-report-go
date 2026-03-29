package repositories

import (
	"context"

	"backend-sport-team-report-go/internal/modules/team/domain/entities"
	"backend-sport-team-report-go/internal/shared/paginator"
)

type TeamRepository interface {
	Create(ctx context.Context, team entities.Team, actorUserID int64) error
	ListByCompany(ctx context.Context, companyID int64, params paginator.Params) (paginator.Result[entities.Team], error)
	FindByIDAndCompany(ctx context.Context, teamID, companyID int64) (entities.Team, error)
	Update(ctx context.Context, team entities.Team, actorUserID int64) error
	SoftDelete(ctx context.Context, teamID, companyID, actorUserID int64) (bool, error)
}
