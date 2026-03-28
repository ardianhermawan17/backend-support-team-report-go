package reports

import (
	"fmt"
	"net/http"
	"testing"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestReportsEnforcesCompanyBoundary(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newReportsRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	tokenCompanyA := createAccountAndLogin(t, accountRepo, router, 7100000000502, 7200000000502, "reports-admin-b", "reports-company-b", "reports-password-b")
	tokenCompanyB := createAccountAndLogin(t, accountRepo, router, 7100000000503, 7200000000503, "reports-admin-c", "reports-company-c", "reports-password-c")

	homeTeamID := createTeamForReportTests(t, router, tokenCompanyA, "Reports Boundary Home FC")
	guestTeamID := createTeamForReportTests(t, router, tokenCompanyA, "Reports Boundary Guest FC")
	topScorerID := createPlayerForReportTests(t, router, tokenCompanyA, homeTeamID, "Reports Boundary Striker", 9)
	scheduleID := createScheduleForReportTests(t, router, tokenCompanyA, homeTeamID, guestTeamID, "2026-08-11", "12:00:00")

	createResponse := sendJSONRequest(t, router, http.MethodPost, "/api/v1/reports", tokenCompanyA, map[string]any{
		"match_schedule_id":           scheduleID,
		"final_score_home":            2,
		"final_score_guest":           1,
		"most_scoring_goal_player_id": topScorerID,
	})
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	createdReport := decodeReportResponse(t, createResponse)

	crossGet := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", createdReport.ID), tokenCompanyB, nil)
	if crossGet.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company get status %d, got %d with body %s", http.StatusNotFound, crossGet.Code, crossGet.Body.String())
	}

	crossDelete := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/reports/%d", createdReport.ID), tokenCompanyB, nil)
	if crossDelete.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company delete status %d, got %d with body %s", http.StatusNotFound, crossDelete.Code, crossDelete.Body.String())
	}
}
