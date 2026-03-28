package domain

import "errors"

var (
	ErrInvalidPlayerInput       = errors.New("invalid player input")
	ErrTeamNotFound             = errors.New("team not found")
	ErrPlayerNotFound           = errors.New("player not found")
	ErrPlayerNumberAlreadyInUse = errors.New("player number already in use")
	ErrPlayerProfileInUse       = errors.New("player profile image already in use")
)
