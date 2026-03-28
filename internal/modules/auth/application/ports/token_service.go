package ports

import "backend-sport-team-report-go/internal/modules/auth/application/dtos"

type TokenService interface {
	Issue(account dtos.AuthenticatedAccount) (dtos.AccessToken, error)
	Parse(token string) (dtos.TokenIdentity, error)
}
