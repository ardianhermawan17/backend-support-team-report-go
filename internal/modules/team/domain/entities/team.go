package entities

import "time"

type Team struct {
	ID                    int64
	CompanyID             int64
	Name                  string
	LogoImageID           *int64
	FoundedYear           int
	HomebaseAddress       string
	CityOfHomebaseAddress string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             *time.Time
}
