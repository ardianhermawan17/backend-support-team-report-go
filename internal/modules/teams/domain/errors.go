package domain

import "errors"

var (
	ErrInvalidTeamInput     = errors.New("invalid team input")
	ErrTeamNotFound         = errors.New("team not found")
	ErrTeamAlreadyExists    = errors.New("team already exists")
	ErrTeamLogoAlreadyInUse = errors.New("team logo already in use")
)
