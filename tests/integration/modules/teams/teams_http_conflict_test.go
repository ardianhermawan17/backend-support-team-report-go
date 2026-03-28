package teams

import (
	"net/http"
	"testing"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestTeamsRejectsDuplicateNameInSameCompany(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newTeamsRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000000204, 7200000000204, "teams-admin-d", "teams-company-d", "teams-password-d")

	firstCreate := sendJSONRequest(t, router, http.MethodPost, "/api/v1/teams", token, map[string]any{
		"name":                     "Duplicate FC",
		"founded_year":             2010,
		"homebase_address":         "Street 19",
		"city_of_homebase_address": "Medan",
	})
	if firstCreate.Code != http.StatusCreated {
		t.Fatalf("expected first create status %d, got %d with body %s", http.StatusCreated, firstCreate.Code, firstCreate.Body.String())
	}

	secondCreate := sendJSONRequest(t, router, http.MethodPost, "/api/v1/teams", token, map[string]any{
		"name":                     "Duplicate FC",
		"founded_year":             2011,
		"homebase_address":         "Street 20",
		"city_of_homebase_address": "Medan",
	})
	if secondCreate.Code != http.StatusConflict {
		t.Fatalf("expected duplicate create status %d, got %d with body %s", http.StatusConflict, secondCreate.Code, secondCreate.Body.String())
	}
}
