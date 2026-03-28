package reports

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ginrouter "backend-sport-team-report-go/internal/api/gin/router"
	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/modules/auth/domain/entities"
	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"
	appcrypto "backend-sport-team-report-go/pkg/crypto"
)

func newReportsRouter(conn *postgres.Connection) http.Handler {
	cfg := config.Config{
		App: config.AppConfig{
			Name: "soccer-team-report",
			Env:  config.EnvTest,
		},
		Database: config.DatabaseConfig{},
		Auth: config.AuthConfig{
			JWTSecret:      "integration-test-secret",
			AccessTokenTTL: 15 * time.Minute,
		},
	}

	return ginrouter.New(cfg, conn, logger.New(cfg.App.Name, cfg.App.Env))
}

func createAccountAndLogin(t *testing.T, repo *authpersistence.AccountRepository, router http.Handler, userID, companyID int64, username, companyName, password string) string {
	t.Helper()

	hash, err := appcrypto.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	account := entities.CompanyAdminAccount{
		User: entities.User{
			ID:           userID,
			Username:     username,
			PasswordHash: hash,
		},
		Company: entities.Company{
			ID:   companyID,
			Name: companyName,
		},
	}

	if err := repo.Create(context.Background(), account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	loginResponse := sendJSONRequest(t, router, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"username": username,
		"password": password,
	})
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected login status %d, got %d with body %s", http.StatusOK, loginResponse.Code, loginResponse.Body.String())
	}

	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	if payload.AccessToken == "" {
		t.Fatal("expected access token from login")
	}

	return payload.AccessToken
}

func createTeamForReportTests(t *testing.T, router http.Handler, token, name string) int64 {
	t.Helper()

	response := sendJSONRequest(t, router, http.MethodPost, "/api/v1/teams", token, map[string]any{
		"name":                     name,
		"founded_year":             2011,
		"homebase_address":         "Jalan Stadion Nomor 12",
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

func createPlayerForReportTests(t *testing.T, router http.Handler, token string, teamID int64, name string, number int) int64 {
	t.Helper()

	response := sendJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/api/v1/teams/%d/players", teamID), token, map[string]any{
		"name":          name,
		"height":        180.5,
		"weight":        75.2,
		"position":      "striker",
		"player_number": number,
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("expected create player status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal create player response: %v", err)
	}
	if payload.ID == 0 {
		t.Fatal("expected player id from create player response")
	}

	return payload.ID
}

func createScheduleForReportTests(t *testing.T, router http.Handler, token string, homeTeamID, guestTeamID int64, date, matchTime string) int64 {
	t.Helper()

	response := sendJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", token, map[string]any{
		"match_date":    date,
		"match_time":    matchTime,
		"home_team_id":  homeTeamID,
		"guest_team_id": guestTeamID,
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("expected create schedule status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal create schedule response: %v", err)
	}
	if payload.ID == 0 {
		t.Fatal("expected schedule id from create schedule response")
	}

	return payload.ID
}

func sendJSONRequest(t *testing.T, router http.Handler, method, path, token string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	return sendRequest(t, router, method, path, token, bytes.NewReader(body))
}

func sendRequest(t *testing.T, router http.Handler, method, path, token string, body *bytes.Reader) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		requestBody = body
	}

	req := httptest.NewRequest(method, path, requestBody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

type reportPayload struct {
	ID                                                               int64  `json:"id"`
	CompanyID                                                        int64  `json:"company_id"`
	MatchScheduleID                                                  int64  `json:"match_schedule_id"`
	HomeTeamID                                                       int64  `json:"home_team_id"`
	GuestTeamID                                                      int64  `json:"guest_team_id"`
	FinalScoreHome                                                   int    `json:"final_score_home"`
	FinalScoreGuest                                                  int    `json:"final_score_guest"`
	StatusMatch                                                      string `json:"status_match"`
	MostScoringGoalPlayerID                                          *int64 `json:"most_scoring_goal_player_id"`
	AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule  int    `json:"accumulate_total_win_for_home_team_from_start_to_the_current_match_schedule"`
	AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule int    `json:"accumulate_total_win_for_guest_team_from_start_to_the_current_match_schedule"`
}

func decodeReportResponse(t *testing.T, response *httptest.ResponseRecorder) reportPayload {
	t.Helper()

	var report reportPayload
	if err := json.Unmarshal(response.Body.Bytes(), &report); err != nil {
		t.Fatalf("unmarshal report response: %v", err)
	}

	return report
}
