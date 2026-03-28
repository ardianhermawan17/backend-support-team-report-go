package requests

type UpsertScheduleRequest struct {
	MatchDate   string `json:"match_date"`
	MatchTime   string `json:"match_time"`
	HomeTeamID  int64  `json:"home_team_id"`
	GuestTeamID int64  `json:"guest_team_id"`
}
