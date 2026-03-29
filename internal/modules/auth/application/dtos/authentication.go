package dtos

import "time"

type AuthenticatedAccount struct {
	UserID      int64
	Username    string
	Email       string
	CompanyID   int64
	CompanyName string
}

type AccessToken struct {
	Value     string
	ExpiresAt time.Time
}

type LoginResult struct {
	Account     AuthenticatedAccount
	AccessToken AccessToken
}

type TokenIdentity struct {
	UserID    int64
	CompanyID int64
	Username  string
	ExpiresAt time.Time
}
