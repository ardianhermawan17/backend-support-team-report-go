package repositories

import (
	"context"

	"backend-sport-team-report-go/internal/modules/auth/domain/entities"
)

type AccountRepository interface {
	Create(ctx context.Context, account entities.CompanyAdminAccount) error
	FindByUsername(ctx context.Context, username string) (entities.CompanyAdminAccount, error)
}
