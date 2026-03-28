package players

import (
	"fmt"
	"net/http"
	"testing"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestPlayersRejectsDuplicateNumberInSameTeam(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newPlayersRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000000304, 7200000000304, "players-admin-d", "players-company-d", "players-password-d")
	teamID := createTeamForPlayerTests(t, router, token, "Players Delta FC")

	firstCreate := sendJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/api/v1/teams/%d/players", teamID), token, map[string]any{
		"name":          "Player One",
		"height":        178.0,
		"weight":        70.0,
		"position":      "defender",
		"player_number": 5,
	})
	if firstCreate.Code != http.StatusCreated {
		t.Fatalf("expected first create status %d, got %d with body %s", http.StatusCreated, firstCreate.Code, firstCreate.Body.String())
	}

	secondCreate := sendJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/api/v1/teams/%d/players", teamID), token, map[string]any{
		"name":          "Player Two",
		"height":        176.0,
		"weight":        68.0,
		"position":      "goalkeeper",
		"player_number": 5,
	})
	if secondCreate.Code != http.StatusConflict {
		t.Fatalf("expected duplicate create status %d, got %d with body %s", http.StatusConflict, secondCreate.Code, secondCreate.Body.String())
	}
}
