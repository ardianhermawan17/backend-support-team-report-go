package repositories

import (
	"context"

	"backend-sport-team-report-go/internal/modules/schedule/domain/entities"
)

type ScheduleRepository interface {
	Create(ctx context.Context, schedule entities.Schedule, actorUserID int64) error
	ListByCompany(ctx context.Context, companyID int64) ([]entities.Schedule, error)
	FindByIDAndCompany(ctx context.Context, scheduleID, companyID int64) (entities.Schedule, error)
	Update(ctx context.Context, schedule entities.Schedule, actorUserID int64) error
	SoftDelete(ctx context.Context, scheduleID, companyID, actorUserID int64) (bool, error)
}
