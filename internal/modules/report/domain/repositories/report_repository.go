package repositories

import (
	"context"

	"backend-sport-team-report-go/internal/modules/report/domain/entities"
)

type ReportRepository interface {
	Create(ctx context.Context, report entities.Report, actorUserID int64) error
	ListByCompany(ctx context.Context, companyID int64) ([]entities.Report, error)
	FindByIDAndCompany(ctx context.Context, reportID, companyID int64) (entities.Report, error)
	Update(ctx context.Context, report entities.Report, actorUserID int64) error
	SoftDelete(ctx context.Context, reportID, companyID, actorUserID int64) (bool, error)
}
