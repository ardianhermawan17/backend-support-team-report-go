package responses

import (
	"time"

	"backend-sport-team-report-go/internal/modules/auth/application/dtos"
)

type AuthAccountResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type AuthCompanyResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type LoginResponse struct {
	AccessToken string              `json:"access_token"`
	TokenType   string              `json:"token_type"`
	ExpiresAt   time.Time           `json:"expires_at"`
	User        AuthAccountResponse `json:"user"`
	Company     AuthCompanyResponse `json:"company"`
}

type CurrentAccountResponse struct {
	User    AuthAccountResponse `json:"user"`
	Company AuthCompanyResponse `json:"company"`
}

func NewLoginResponse(result dtos.LoginResult) LoginResponse {
	return LoginResponse{
		AccessToken: result.AccessToken.Value,
		TokenType:   "Bearer",
		ExpiresAt:   result.AccessToken.ExpiresAt,
		User:        newAccountResponse(result.Account),
		Company:     newCompanyResponse(result.Account),
	}
}

func NewCurrentAccountResponse(account dtos.AuthenticatedAccount) CurrentAccountResponse {
	return CurrentAccountResponse{
		User:    newAccountResponse(account),
		Company: newCompanyResponse(account),
	}
}

func newAccountResponse(account dtos.AuthenticatedAccount) AuthAccountResponse {
	return AuthAccountResponse{
		ID:       account.UserID,
		Username: account.Username,
		Email:    account.Email,
	}
}

func newCompanyResponse(account dtos.AuthenticatedAccount) AuthCompanyResponse {
	return AuthCompanyResponse{
		ID:   account.CompanyID,
		Name: account.CompanyName,
	}
}
