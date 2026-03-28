package domain

import "errors"

var (
	ErrInvalidScheduleInput  = errors.New("invalid schedule input")
	ErrScheduleNotFound      = errors.New("schedule not found")
	ErrScheduleAlreadyExists = errors.New("schedule already exists")
	ErrScheduleTeamNotFound  = errors.New("schedule team not found")
)
