package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"backend-sport-team-report-go/internal/modules/auth/application/dtos"
	authdomain "backend-sport-team-report-go/internal/modules/auth/domain"
	"backend-sport-team-report-go/internal/modules/auth/domain/entities"
	"backend-sport-team-report-go/internal/modules/auth/domain/repositories"
)

type CurrentAccountHandler struct {
	accounts repositories.AccountRepository
}

func NewCurrentAccountHandler(accounts repositories.AccountRepository) CurrentAccountHandler {
	return CurrentAccountHandler{accounts: accounts}
}

func (h CurrentAccountHandler) Handle(ctx context.Context, identity dtos.TokenIdentity) (dtos.AuthenticatedAccount, error) {
	account, err := h.accounts.FindByUsername(ctx, identity.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dtos.AuthenticatedAccount{}, authdomain.ErrUnauthorized
		}
		return dtos.AuthenticatedAccount{}, fmt.Errorf("find current account: %w", err)
	}

	if account.User.ID != identity.UserID || account.Company.ID != identity.CompanyID {
		return dtos.AuthenticatedAccount{}, authdomain.ErrUnauthorized
	}

	return toAuthenticatedAccount(account), nil
}

func toAuthenticatedAccount(account entities.CompanyAdminAccount) dtos.AuthenticatedAccount {
	return dtos.AuthenticatedAccount{
		UserID:      account.User.ID,
		Username:    account.User.Username,
		Email:       account.User.Email,
		CompanyID:   account.Company.ID,
		CompanyName: account.Company.Name,
	}
}
