package repositories

import (
	"context"

	"backend-sport-team-report-go/internal/modules/report/domain/entities"
	"backend-sport-team-report-go/internal/shared/paginator"
)

type ReportRepository interface {
	Create(ctx context.Context, report entities.Report, actorUserID int64) error
	ListByCompany(ctx context.Context, companyID int64, params paginator.Params) (paginator.Result[entities.Report], error)
	FindByIDAndCompany(ctx context.Context, reportID, companyID int64) (entities.Report, error)
	Update(ctx context.Context, report entities.Report, actorUserID int64) error
	SoftDelete(ctx context.Context, reportID, companyID, actorUserID int64) (bool, error)
}
