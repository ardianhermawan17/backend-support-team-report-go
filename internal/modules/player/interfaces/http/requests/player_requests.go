package requests

type UpsertPlayerRequest struct {
	Name           string  `json:"name"`
	Height         float64 `json:"height"`
	Weight         float64 `json:"weight"`
	Position       string  `json:"position"`
	PlayerNumber   int     `json:"player_number"`
	ProfileImageID *int64  `json:"profile_image_id"`
}
