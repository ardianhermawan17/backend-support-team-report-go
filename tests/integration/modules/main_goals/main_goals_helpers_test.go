package main_goals

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	ginrouter "backend-sport-team-report-go/internal/api/gin/router"
	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	appseeding "backend-sport-team-report-go/internal/platform/database/seeding"
	"backend-sport-team-report-go/internal/platform/idgenerator"
	"backend-sport-team-report-go/internal/shared/logger"
	"backend-sport-team-report-go/tests/integration/testhelpers"
)

const (
	mainGoalUsername    = "main.goal"
	mainGoalEmail       = "main.goal@gmail.com"
	mainGoalPassword    = "password"
	mainGoalCompanyName = "Main Goal Soccer Co."
)

var (
	mainGoalsUserIDs    = mustNewSnowflakeGenerator(5)
	mainGoalsCompanyIDs = mustNewSnowflakeGenerator(6)
)

type scenarioSession struct {
	Token     string
	UserID    int64
	CompanyID int64
}

type teamPayload struct {
	ID                    int64  `json:"id"`
	CompanyID             int64  `json:"company_id"`
	Name                  string `json:"name"`
	LogoImageID           *int64 `json:"logo_image_id"`
	FoundedYear           int    `json:"founded_year"`
	HomebaseAddress       string `json:"homebase_address"`
	CityOfHomebaseAddress string `json:"city_of_homebase_address"`
}

type playerPayload struct {
	ID             int64   `json:"id"`
	TeamID         int64   `json:"team_id"`
	Name           string  `json:"name"`
	Height         float64 `json:"height"`
	Weight         float64 `json:"weight"`
	Position       string  `json:"position"`
	PlayerNumber   int     `json:"player_number"`
	ProfileImageID *int64  `json:"profile_image_id"`
}

type schedulePayload struct {
	ID          int64  `json:"id"`
	CompanyID   int64  `json:"company_id"`
	MatchDate   string `json:"match_date"`
	MatchTime   string `json:"match_time"`
	HomeTeamID  int64  `json:"home_team_id"`
	GuestTeamID int64  `json:"guest_team_id"`
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

type itemsResponse[T any] struct {
	Items []T `json:"items"`
}

func newMainGoalsRouter(conn *postgres.Connection) http.Handler {
	cfg := testhelpers.DefaultTestConfig()
	return ginrouter.New(cfg, conn, logger.New(cfg.App.Name, cfg.App.Env))
}

func seedMainGoalAndLogin(t *testing.T, db *sql.DB, router http.Handler) scenarioSession {
	t.Helper()

	service, err := appseeding.NewService(db)
	if err != nil {
		t.Fatalf("create seeding service: %v", err)
	}

	if err := service.Seed(context.Background()); err != nil {
		t.Fatalf("seed bootstrap data: %v", err)
	}

	if err := service.SeedMainGoalScenarioAccount(context.Background()); err != nil {
		t.Fatalf("seed main.goal scenario account: %v", err)
	}

	response := sendJSONRequest(t, router, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"username": mainGoalUsername,
		"password": mainGoalPassword,
	})
	if response.Code != http.StatusOK {
		t.Fatalf("expected main.goal login status %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		User        struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"user"`
		Company struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"company"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal main.goal login response: %v", err)
	}
	if payload.AccessToken == "" {
		t.Fatal("expected access_token from main.goal login")
	}
	if payload.User.Username != mainGoalUsername || payload.User.Email != mainGoalEmail {
		t.Fatalf("unexpected main.goal user payload: %#v", payload.User)
	}
	if payload.Company.Name != mainGoalCompanyName {
		t.Fatalf("expected main.goal company %q, got %q", mainGoalCompanyName, payload.Company.Name)
	}

	return scenarioSession{
		Token:     payload.AccessToken,
		UserID:    payload.User.ID,
		CompanyID: payload.Company.ID,
	}
}

func createScenarioAccountAndLogin(t *testing.T, repo *authpersistence.AccountRepository, router http.Handler, username, companyName, password string) scenarioSession {
	t.Helper()

	userID, err := mainGoalsUserIDs.NewID()
	if err != nil {
		t.Fatalf("generate scenario user id: %v", err)
	}

	companyID, err := mainGoalsCompanyIDs.NewID()
	if err != nil {
		t.Fatalf("generate scenario company id: %v", err)
	}

	return scenarioSession{
		Token:     testhelpers.CreateAccountAndLogin(t, repo, router, userID, companyID, username, companyName, password),
		UserID:    userID,
		CompanyID: companyID,
	}
}

func createTeamForScenario(t *testing.T, router http.Handler, token, name string, foundedYear int, homebaseAddress, city string) teamPayload {
	t.Helper()

	response := sendJSONRequest(t, router, http.MethodPost, "/api/v1/teams", token, map[string]any{
		"name":                     name,
		"founded_year":             foundedYear,
		"homebase_address":         homebaseAddress,
		"city_of_homebase_address": city,
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("expected create team status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
	}

	team := decodeJSONResponse[teamPayload](t, response)
	if team.ID == 0 {
		t.Fatal("expected team id from create team response")
	}

	return team
}

func createPlayerForScenario(t *testing.T, router http.Handler, token string, teamID int64, name string, height, weight float64, position string, playerNumber int) playerPayload {
	t.Helper()

	response := sendJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/api/v1/teams/%d/players", teamID), token, map[string]any{
		"name":          name,
		"height":        height,
		"weight":        weight,
		"position":      position,
		"player_number": playerNumber,
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("expected create player status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
	}

	player := decodeJSONResponse[playerPayload](t, response)
	if player.ID == 0 {
		t.Fatal("expected player id from create player response")
	}

	return player
}

func createScheduleForScenario(t *testing.T, router http.Handler, token, matchDate, matchTime string, homeTeamID, guestTeamID int64) schedulePayload {
	t.Helper()

	response := sendJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", token, map[string]any{
		"match_date":    matchDate,
		"match_time":    matchTime,
		"home_team_id":  homeTeamID,
		"guest_team_id": guestTeamID,
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("expected create schedule status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
	}

	schedule := decodeJSONResponse[schedulePayload](t, response)
	if schedule.ID == 0 {
		t.Fatal("expected schedule id from create schedule response")
	}

	return schedule
}

func createReportForScenario(t *testing.T, router http.Handler, token string, matchScheduleID int64, finalScoreHome, finalScoreGuest int, mostScoringGoalPlayerID *int64) reportPayload {
	t.Helper()

	response := sendJSONRequest(t, router, http.MethodPost, "/api/v1/reports", token, map[string]any{
		"match_schedule_id":           matchScheduleID,
		"final_score_home":            finalScoreHome,
		"final_score_guest":           finalScoreGuest,
		"most_scoring_goal_player_id": mostScoringGoalPlayerID,
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("expected create report status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
	}

	report := decodeJSONResponse[reportPayload](t, response)
	if report.ID == 0 {
		t.Fatal("expected report id from create report response")
	}

	return report
}

func sendJSONRequest(t *testing.T, router http.Handler, method, path, token string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	return testhelpers.SendJSONRequest(t, router, method, path, token, payload)
}

func sendRequest(t *testing.T, router http.Handler, method, path, token string, body *bytes.Reader) *httptest.ResponseRecorder {
	t.Helper()
	return testhelpers.SendRequest(t, router, method, path, token, body)
}

func decodeJSONResponse[T any](t *testing.T, response *httptest.ResponseRecorder) T {
	t.Helper()

	var payload T
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}

	return payload
}

func assertSoftDeletedAt(t *testing.T, db *sql.DB, table string, id int64) {
	t.Helper()

	var deletedAt sql.NullTime
	query := fmt.Sprintf("SELECT deleted_at FROM %s WHERE id = $1", table)
	if err := db.QueryRow(query, id).Scan(&deletedAt); err != nil {
		t.Fatalf("query deleted_at from %s for %d: %v", table, id, err)
	}
	if !deletedAt.Valid {
		t.Fatalf("expected %s deleted_at to be set for id %d", table, id)
	}
}

func mustNewSnowflakeGenerator(nodeID int64) *idgenerator.SnowflakeGenerator {
	generator, err := idgenerator.NewSnowflakeGenerator(nodeID)
	if err != nil {
		panic(err)
	}

	return generator
}
