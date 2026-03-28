package responses

import (
	"backend-sport-team-report-go/internal/modules/teams/domain/entities"
)

type TeamResponse struct {
	ID                    int64  `json:"id"`
	CompanyID             int64  `json:"company_id"`
	Name                  string `json:"name"`
	LogoImageID           *int64 `json:"logo_image_id"`
	FoundedYear           int    `json:"founded_year"`
	HomebaseAddress       string `json:"homebase_address"`
	CityOfHomebaseAddress string `json:"city_of_homebase_address"`
}

func NewTeamResponse(team entities.Team) TeamResponse {
	return TeamResponse{
		ID:                    team.ID,
		CompanyID:             team.CompanyID,
		Name:                  team.Name,
		LogoImageID:           team.LogoImageID,
		FoundedYear:           team.FoundedYear,
		HomebaseAddress:       team.HomebaseAddress,
		CityOfHomebaseAddress: team.CityOfHomebaseAddress,
	}
}

func NewTeamListResponse(teams []entities.Team) []TeamResponse {
	responses := make([]TeamResponse, 0, len(teams))
	for _, team := range teams {
		responses = append(responses, NewTeamResponse(team))
	}

	return responses
}
