package responses

import "backend-sport-team-report-go/internal/modules/player/domain/entities"

type PlayerResponse struct {
	ID             int64   `json:"id"`
	TeamID         int64   `json:"team_id"`
	Name           string  `json:"name"`
	Height         float64 `json:"height"`
	Weight         float64 `json:"weight"`
	Position       string  `json:"position"`
	PlayerNumber   int     `json:"player_number"`
	ProfileImageID *int64  `json:"profile_image_id"`
}

func NewPlayerResponse(player entities.Player) PlayerResponse {
	return PlayerResponse{
		ID:             player.ID,
		TeamID:         player.TeamID,
		Name:           player.Name,
		Height:         player.Height,
		Weight:         player.Weight,
		Position:       player.Position,
		PlayerNumber:   player.PlayerNumber,
		ProfileImageID: player.ProfileImageID,
	}
}

func NewPlayerListResponse(players []entities.Player) []PlayerResponse {
	responses := make([]PlayerResponse, 0, len(players))
	for _, player := range players {
		responses = append(responses, NewPlayerResponse(player))
	}

	return responses
}
