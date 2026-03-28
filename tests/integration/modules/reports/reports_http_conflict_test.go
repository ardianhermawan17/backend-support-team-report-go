package reports

import (
	"net/http"
	"testing"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestReportsRejectDuplicateScheduleReport(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newReportsRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000000504, 7200000000504, "reports-admin-d", "reports-company-d", "reports-password-d")
	homeTeamID := createTeamForReportTests(t, router, token, "Reports Conflict Home FC")
	guestTeamID := createTeamForReportTests(t, router, token, "Reports Conflict Guest FC")
	topScorerID := createPlayerForReportTests(t, router, token, homeTeamID, "Reports Conflict Striker", 11)
	scheduleID := createScheduleForReportTests(t, router, token, homeTeamID, guestTeamID, "2026-08-12", "09:30:00")

	firstCreate := sendJSONRequest(t, router, http.MethodPost, "/api/v1/reports", token, map[string]any{
		"match_schedule_id":           scheduleID,
		"final_score_home":            1,
		"final_score_guest":           0,
		"most_scoring_goal_player_id": topScorerID,
	})
	if firstCreate.Code != http.StatusCreated {
		t.Fatalf("expected first create status %d, got %d with body %s", http.StatusCreated, firstCreate.Code, firstCreate.Body.String())
	}

	duplicateCreate := sendJSONRequest(t, router, http.MethodPost, "/api/v1/reports", token, map[string]any{
		"match_schedule_id":           scheduleID,
		"final_score_home":            2,
		"final_score_guest":           1,
		"most_scoring_goal_player_id": topScorerID,
	})
	if duplicateCreate.Code != http.StatusConflict {
		t.Fatalf("expected duplicate create status %d, got %d with body %s", http.StatusConflict, duplicateCreate.Code, duplicateCreate.Body.String())
	}
}

func TestReportsRejectTopScorerFromOutsideMatchTeams(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newReportsRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000000505, 7200000000505, "reports-admin-e", "reports-company-e", "reports-password-e")
	homeTeamID := createTeamForReportTests(t, router, token, "Reports Outside Home FC")
	guestTeamID := createTeamForReportTests(t, router, token, "Reports Outside Guest FC")
	outsideTeamID := createTeamForReportTests(t, router, token, "Reports Outside Third FC")
	outsidePlayerID := createPlayerForReportTests(t, router, token, outsideTeamID, "Reports Outside Striker", 7)
	scheduleID := createScheduleForReportTests(t, router, token, homeTeamID, guestTeamID, "2026-08-13", "18:10:00")

	createResponse := sendJSONRequest(t, router, http.MethodPost, "/api/v1/reports", token, map[string]any{
		"match_schedule_id":           scheduleID,
		"final_score_home":            4,
		"final_score_guest":           3,
		"most_scoring_goal_player_id": outsidePlayerID,
	})
	if createResponse.Code != http.StatusNotFound {
		t.Fatalf("expected outside-top-scorer status %d, got %d with body %s", http.StatusNotFound, createResponse.Code, createResponse.Body.String())
	}
}
