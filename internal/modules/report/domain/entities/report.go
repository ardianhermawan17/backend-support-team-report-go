package entities

import "time"

type Report struct {
	ID        int64
	CompanyID int64

	MatchScheduleID int64
	HomeTeamID      int64
	GuestTeamID     int64

	FinalScoreHome  int
	FinalScoreGuest int
	StatusMatch     string

	MostScoringGoalPlayerID *int64

	AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule  int
	AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule int

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
