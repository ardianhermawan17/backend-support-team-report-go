package teams

import (
	"fmt"
	"net/http"
	"testing"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestTeamsEnforcesCompanyBoundary(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newTeamsRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	tokenCompanyA := createAccountAndLogin(t, accountRepo, router, 7100000000202, 7200000000202, "teams-admin-b", "teams-company-b", "teams-password-b")
	tokenCompanyB := createAccountAndLogin(t, accountRepo, router, 7100000000203, 7200000000203, "teams-admin-c", "teams-company-c", "teams-password-c")

	createResponse := sendJSONRequest(t, router, http.MethodPost, "/api/v1/teams", tokenCompanyA, map[string]any{
		"name":                     "Boundary FC",
		"founded_year":             2008,
		"homebase_address":         "Street 12",
		"city_of_homebase_address": "Surabaya",
	})
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	createdTeam := decodeTeamResponse(t, createResponse)

	crossGet := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d", createdTeam.ID), tokenCompanyB, nil)
	if crossGet.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company get status %d, got %d with body %s", http.StatusNotFound, crossGet.Code, crossGet.Body.String())
	}

	crossDelete := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/teams/%d", createdTeam.ID), tokenCompanyB, nil)
	if crossDelete.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company delete status %d, got %d with body %s", http.StatusNotFound, crossDelete.Code, crossDelete.Body.String())
	}
}
