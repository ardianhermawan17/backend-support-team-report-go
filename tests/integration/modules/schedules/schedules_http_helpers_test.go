package schedules

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ginrouter "backend-sport-team-report-go/internal/api/gin/router"
	"backend-sport-team-report-go/internal/config"
	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"
	"backend-sport-team-report-go/tests/integration/testhelpers"
)

func newSchedulesRouter(conn *postgres.Connection) http.Handler {
	cfg := testhelpers.DefaultTestConfig()
	return ginrouter.New(cfg, conn, logger.New(cfg.App.Name, cfg.App.Env))
}

func newSchedulesRouterWithSecurity(conn *postgres.Connection, security config.SecurityConfig) http.Handler {
	cfg := testhelpers.DefaultTestConfig()
	cfg.Security = security
	return ginrouter.New(cfg, conn, logger.New(cfg.App.Name, cfg.App.Env))
}

func createAccountAndLogin(t *testing.T, repo *authpersistence.AccountRepository, router http.Handler, userID, companyID int64, username, companyName, password string) string {
	t.Helper()
	return testhelpers.CreateAccountAndLogin(t, repo, router, userID, companyID, username, companyName, password)
}

func createTeamForScheduleTests(t *testing.T, router http.Handler, token, name string) int64 {
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

type schedulePayload struct {
	ID          int64  `json:"id"`
	CompanyID   int64  `json:"company_id"`
	MatchDate   string `json:"match_date"`
	MatchTime   string `json:"match_time"`
	HomeTeamID  int64  `json:"home_team_id"`
	GuestTeamID int64  `json:"guest_team_id"`
}

func decodeScheduleResponse(t *testing.T, response *httptest.ResponseRecorder) schedulePayload {
	t.Helper()

	var schedule schedulePayload
	if err := json.Unmarshal(response.Body.Bytes(), &schedule); err != nil {
		t.Fatalf("unmarshal schedule response: %v", err)
	}

	return schedule
}
