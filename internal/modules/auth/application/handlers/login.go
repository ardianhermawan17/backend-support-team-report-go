package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"backend-sport-team-report-go/internal/modules/auth/application/dtos"
	"backend-sport-team-report-go/internal/modules/auth/application/ports"
	authdomain "backend-sport-team-report-go/internal/modules/auth/domain"
	"backend-sport-team-report-go/internal/modules/auth/domain/repositories"
	appcrypto "backend-sport-team-report-go/pkg/crypto"
)

type LoginHandler struct {
	accounts repositories.AccountRepository
	tokens   ports.TokenService
}

func NewLoginHandler(accounts repositories.AccountRepository, tokens ports.TokenService) LoginHandler {
	return LoginHandler{accounts: accounts, tokens: tokens}
}

func (h LoginHandler) Handle(ctx context.Context, username, password string) (dtos.LoginResult, error) {
	account, err := h.accounts.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dtos.LoginResult{}, authdomain.ErrInvalidCredentials
		}
		return dtos.LoginResult{}, fmt.Errorf("find account for login: %w", err)
	}

	if err := appcrypto.VerifyPassword(password, account.User.PasswordHash); err != nil {
		return dtos.LoginResult{}, authdomain.ErrInvalidCredentials
	}

	authenticatedAccount := toAuthenticatedAccount(account)
	accessToken, err := h.tokens.Issue(authenticatedAccount)
	if err != nil {
		return dtos.LoginResult{}, fmt.Errorf("issue access token: %w", err)
	}

	return dtos.LoginResult{
		Account:     authenticatedAccount,
		AccessToken: accessToken,
	}, nil
}
