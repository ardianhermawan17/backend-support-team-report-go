package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/modules/auth/application/dtos"
	"backend-sport-team-report-go/internal/modules/auth/application/ports"
)

type TokenService struct {
	secret []byte
	ttl    time.Duration
}

type accessTokenClaims struct {
	Username  string `json:"username"`
	UserID    int64  `json:"user_id"`
	CompanyID int64  `json:"company_id"`
	jwt.RegisteredClaims
}

func NewTokenService(cfg config.AuthConfig) ports.TokenService {
	return TokenService{
		secret: []byte(cfg.JWTSecret),
		ttl:    cfg.AccessTokenTTL,
	}
}

func (s TokenService) Issue(account dtos.AuthenticatedAccount) (dtos.AccessToken, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.ttl)

	claims := accessTokenClaims{
		Username:  account.Username,
		UserID:    account.UserID,
		CompanyID: account.CompanyID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   account.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return dtos.AccessToken{}, fmt.Errorf("sign token: %w", err)
	}

	return dtos.AccessToken{Value: signedToken, ExpiresAt: expiresAt}, nil
}

func (s TokenService) Parse(token string) (dtos.TokenIdentity, error) {
	claims := accessTokenClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return s.secret, nil
	})
	if err != nil {
		return dtos.TokenIdentity{}, fmt.Errorf("parse token: %w", err)
	}

	if !parsedToken.Valid || claims.ExpiresAt == nil {
		return dtos.TokenIdentity{}, fmt.Errorf("parse token: invalid token")
	}

	return dtos.TokenIdentity{
		UserID:    claims.UserID,
		CompanyID: claims.CompanyID,
		Username:  claims.Username,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}
