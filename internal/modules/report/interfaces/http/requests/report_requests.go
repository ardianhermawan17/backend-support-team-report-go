package requests

type UpsertReportRequest struct {
	MatchScheduleID         int64  `json:"match_schedule_id"`
	FinalScoreHome          int    `json:"final_score_home"`
	FinalScoreGuest         int    `json:"final_score_guest"`
	MostScoringGoalPlayerID *int64 `json:"most_scoring_goal_player_id"`
}
