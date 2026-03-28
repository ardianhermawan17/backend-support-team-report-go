package requests

type UpsertTeamRequest struct {
	Name                  string `json:"name"`
	LogoImageID           *int64 `json:"logo_image_id"`
	FoundedYear           int    `json:"founded_year"`
	HomebaseAddress       string `json:"homebase_address"`
	CityOfHomebaseAddress string `json:"city_of_homebase_address"`
}
