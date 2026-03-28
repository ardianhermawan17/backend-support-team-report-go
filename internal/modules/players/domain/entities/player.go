package entities

import "time"

type Player struct {
	ID             int64
	TeamID         int64
	Name           string
	Height         float64
	Weight         float64
	Position       string
	PlayerNumber   int
	ProfileImageID *int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}
