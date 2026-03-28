package entities

import "time"

type Schedule struct {
	ID          int64
	CompanyID   int64
	MatchDate   time.Time
	MatchTime   time.Time
	HomeTeamID  int64
	GuestTeamID int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
