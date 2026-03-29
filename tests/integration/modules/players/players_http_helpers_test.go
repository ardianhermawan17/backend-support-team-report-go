package players

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ginrouter "backend-sport-team-report-go/internal/api/gin/router"
	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"
	"backend-sport-team-report-go/tests/integration/testhelpers"
)

func newPlayersRouter(conn *postgres.Connection) http.Handler {
	cfg := testhelpers.DefaultTestConfig()
	return ginrouter.New(cfg, conn, logger.New(cfg.App.Name, cfg.App.Env))
}

func createAccountAndLogin(t *testing.T, repo *authpersistence.AccountRepository, router http.Handler, userID, companyID int64, username, companyName, password string) string {
	t.Helper()
	return testhelpers.CreateAccountAndLogin(t, repo, router, userID, companyID, username, companyName, password)
}

func createTeamForPlayerTests(t *testing.T, router http.Handler, token, name string) int64 {
	t.Helper()

	response := sendJSONRequest(t, router, http.MethodPost, "/api/v1/teams", token, map[string]any{
		"name":                     name,
		"founded_year":             2013,
		"homebase_address":         "Jalan Lapangan Nomor 7",
		"city_of_homebase_address": "Bandung",
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("expected create team status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal create team response: %v", err)
	}
	if payload.ID == 0 {
		t.Fatal("expected team id from create team response")
	}

	return payload.ID
}

func sendJSONRequest(t *testing.T, router http.Handler, method, path, token string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	return testhelpers.SendJSONRequest(t, router, method, path, token, payload)
}

func sendRequest(t *testing.T, router http.Handler, method, path, token string, body *bytes.Reader) *httptest.ResponseRecorder {
	t.Helper()
	return testhelpers.SendRequest(t, router, method, path, token, body)
}

type playerPayload struct {
	ID           int64   `json:"id"`
	TeamID       int64   `json:"team_id"`
	Name         string  `json:"name"`
	Height       float64 `json:"height"`
	Weight       float64 `json:"weight"`
	Position     string  `json:"position"`
	PlayerNumber int     `json:"player_number"`
}

func decodePlayerResponse(t *testing.T, response *httptest.ResponseRecorder) playerPayload {
	t.Helper()

	var player playerPayload
	if err := json.Unmarshal(response.Body.Bytes(), &player); err != nil {
		t.Fatalf("unmarshal player response: %v", err)
	}

	return player
}
