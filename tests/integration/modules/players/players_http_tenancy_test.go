package players

import (
	"fmt"
	"net/http"
	"testing"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestPlayersEnforcesCompanyBoundary(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newPlayersRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	tokenCompanyA := createAccountAndLogin(t, accountRepo, router, 7100000000302, 7200000000302, "players-admin-b", "players-company-b", "players-password-b")
	tokenCompanyB := createAccountAndLogin(t, accountRepo, router, 7100000000303, 7200000000303, "players-admin-c", "players-company-c", "players-password-c")

	teamID := createTeamForPlayerTests(t, router, tokenCompanyA, "Players Boundary FC")

	createResponse := sendJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/api/v1/teams/%d/players", teamID), tokenCompanyA, map[string]any{
		"name":          "Boundary Player",
		"height":        182.0,
		"weight":        78.0,
		"position":      "striker",
		"player_number": 11,
	})
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	createdPlayer := decodePlayerResponse(t, createResponse)

	crossGet := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d/players/%d", teamID, createdPlayer.ID), tokenCompanyB, nil)
	if crossGet.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company get status %d, got %d with body %s", http.StatusNotFound, crossGet.Code, crossGet.Body.String())
	}

	crossDelete := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/teams/%d/players/%d", teamID, createdPlayer.ID), tokenCompanyB, nil)
	if crossDelete.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company delete status %d, got %d with body %s", http.StatusNotFound, crossDelete.Code, crossDelete.Body.String())
	}
}
