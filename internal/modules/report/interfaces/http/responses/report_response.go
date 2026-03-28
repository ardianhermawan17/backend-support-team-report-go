package responses

import "backend-sport-team-report-go/internal/modules/report/domain/entities"

type ReportResponse struct {
	ID                                                               int64  `json:"id"`
	CompanyID                                                        int64  `json:"company_id"`
	MatchScheduleID                                                  int64  `json:"match_schedule_id"`
	HomeTeamID                                                       int64  `json:"home_team_id"`
	GuestTeamID                                                      int64  `json:"guest_team_id"`
	FinalScoreHome                                                   int    `json:"final_score_home"`
	FinalScoreGuest                                                  int    `json:"final_score_guest"`
	StatusMatch                                                      string `json:"status_match"`
	MostScoringGoalPlayerID                                          *int64 `json:"most_scoring_goal_player_id"`
	AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule  int    `json:"accumulate_total_win_for_home_team_from_start_to_the_current_match_schedule"`
	AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule int    `json:"accumulate_total_win_for_guest_team_from_start_to_the_current_match_schedule"`
}

func NewReportResponse(report entities.Report) ReportResponse {
	return ReportResponse{
		ID:                      report.ID,
		CompanyID:               report.CompanyID,
		MatchScheduleID:         report.MatchScheduleID,
		HomeTeamID:              report.HomeTeamID,
		GuestTeamID:             report.GuestTeamID,
		FinalScoreHome:          report.FinalScoreHome,
		FinalScoreGuest:         report.FinalScoreGuest,
		StatusMatch:             report.StatusMatch,
		MostScoringGoalPlayerID: report.MostScoringGoalPlayerID,
		AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule:  report.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule,
		AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule: report.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule,
	}
}

func NewReportListResponse(reports []entities.Report) []ReportResponse {
	responses := make([]ReportResponse, 0, len(reports))
	for _, report := range reports {
		responses = append(responses, NewReportResponse(report))
	}

	return responses
}
