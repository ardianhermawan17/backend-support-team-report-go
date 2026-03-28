package reports

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestReportsCRUDWithinCompany(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newReportsRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000000501, 7200000000501, "reports-admin-a", "reports-company-a", "reports-password-a")
	homeTeamID := createTeamForReportTests(t, router, token, "Reports Home FC")
	guestTeamID := createTeamForReportTests(t, router, token, "Reports Guest FC")
	topScorerID := createPlayerForReportTests(t, router, token, homeTeamID, "Reports Striker", 10)
	scheduleID := createScheduleForReportTests(t, router, token, homeTeamID, guestTeamID, "2026-08-10", "15:30:00")

	createResponse := sendJSONRequest(t, router, http.MethodPost, "/api/v1/reports", token, map[string]any{
		"match_schedule_id":           scheduleID,
		"final_score_home":            3,
		"final_score_guest":           1,
		"most_scoring_goal_player_id": topScorerID,
	})
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	createdReport := decodeReportResponse(t, createResponse)
	if createdReport.StatusMatch != "home_team_win" {
		t.Fatalf("expected status_match home_team_win, got %q", createdReport.StatusMatch)
	}
	if createdReport.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule != 1 {
		t.Fatalf("expected home accumulated wins 1, got %d", createdReport.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule)
	}
	if createdReport.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule != 0 {
		t.Fatalf("expected guest accumulated wins 0, got %d", createdReport.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule)
	}

	getResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", createdReport.ID), token, nil)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected get status %d, got %d with body %s", http.StatusOK, getResponse.Code, getResponse.Body.String())
	}

	listResponse := sendRequest(t, router, http.MethodGet, "/api/v1/reports", token, nil)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, listResponse.Code, listResponse.Body.String())
	}

	var listPayload struct {
		Items []struct {
			ID int64 `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listResponse.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("unmarshal list response: %v", err)
	}
	if len(listPayload.Items) != 1 || listPayload.Items[0].ID != createdReport.ID {
		t.Fatalf("expected one report %d in list, got %#v", createdReport.ID, listPayload.Items)
	}

	updateResponse := sendJSONRequest(t, router, http.MethodPut, fmt.Sprintf("/api/v1/reports/%d", createdReport.ID), token, map[string]any{
		"match_schedule_id":           scheduleID,
		"final_score_home":            0,
		"final_score_guest":           0,
		"most_scoring_goal_player_id": nil,
	})
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d with body %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}

	updatedReport := decodeReportResponse(t, updateResponse)
	if updatedReport.StatusMatch != "draw" {
		t.Fatalf("expected status_match draw, got %q", updatedReport.StatusMatch)
	}
	if updatedReport.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule != 0 {
		t.Fatalf("expected home accumulated wins 0 after draw update, got %d", updatedReport.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule)
	}

	deleteResponse := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/reports/%d", updatedReport.ID), token, nil)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected delete status %d, got %d with body %s", http.StatusNoContent, deleteResponse.Code, deleteResponse.Body.String())
	}

	getDeletedResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", updatedReport.ID), token, nil)
	if getDeletedResponse.Code != http.StatusNotFound {
		t.Fatalf("expected get deleted status %d, got %d with body %s", http.StatusNotFound, getDeletedResponse.Code, getDeletedResponse.Body.String())
	}

	var deletedAt *time.Time
	if err := env.DB.QueryRow(`SELECT deleted_at FROM reports WHERE id = $1`, updatedReport.ID).Scan(&deletedAt); err != nil {
		t.Fatalf("query deleted_at: %v", err)
	}
	if deletedAt == nil {
		t.Fatal("expected deleted_at to be set after delete")
	}
}
