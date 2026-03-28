package responses

import (
	"backend-sport-team-report-go/internal/modules/schedule/domain/entities"
)

type ScheduleResponse struct {
	ID          int64  `json:"id"`
	CompanyID   int64  `json:"company_id"`
	MatchDate   string `json:"match_date"`
	MatchTime   string `json:"match_time"`
	HomeTeamID  int64  `json:"home_team_id"`
	GuestTeamID int64  `json:"guest_team_id"`
}

func NewScheduleResponse(schedule entities.Schedule) ScheduleResponse {
	return ScheduleResponse{
		ID:          schedule.ID,
		CompanyID:   schedule.CompanyID,
		MatchDate:   schedule.MatchDate.Format("2006-01-02"),
		MatchTime:   schedule.MatchTime.Format("15:04:05"),
		HomeTeamID:  schedule.HomeTeamID,
		GuestTeamID: schedule.GuestTeamID,
	}
}

func NewScheduleListResponse(schedules []entities.Schedule) []ScheduleResponse {
	responses := make([]ScheduleResponse, 0, len(schedules))
	for _, schedule := range schedules {
		responses = append(responses, NewScheduleResponse(schedule))
	}

	return responses
}
