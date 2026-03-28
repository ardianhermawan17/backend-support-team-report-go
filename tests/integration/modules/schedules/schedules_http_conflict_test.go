package schedules

import (
	"net/http"
	"testing"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestSchedulesRejectsDuplicateIdentityInSameCompany(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newSchedulesRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000000404, 7200000000404, "schedules-admin-d", "schedules-company-d", "schedules-password-d")
	homeTeamID := createTeamForScheduleTests(t, router, token, "Schedules Conflict Home FC")
	guestTeamID := createTeamForScheduleTests(t, router, token, "Schedules Conflict Guest FC")

	firstCreate := sendJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", token, map[string]any{
		"match_date":    "2026-07-13",
		"match_time":    "11:00:00",
		"home_team_id":  homeTeamID,
		"guest_team_id": guestTeamID,
	})
	if firstCreate.Code != http.StatusCreated {
		t.Fatalf("expected first create status %d, got %d with body %s", http.StatusCreated, firstCreate.Code, firstCreate.Body.String())
	}

	duplicateCreate := sendJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", token, map[string]any{
		"match_date":    "2026-07-13",
		"match_time":    "11:00:00",
		"home_team_id":  homeTeamID,
		"guest_team_id": guestTeamID,
	})
	if duplicateCreate.Code != http.StatusConflict {
		t.Fatalf("expected duplicate create status %d, got %d with body %s", http.StatusConflict, duplicateCreate.Code, duplicateCreate.Body.String())
	}
}

func TestSchedulesRejectsSameHomeAndGuestTeam(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newSchedulesRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000000405, 7200000000405, "schedules-admin-e", "schedules-company-e", "schedules-password-e")
	homeTeamID := createTeamForScheduleTests(t, router, token, "Schedules Same Team FC")

	createResponse := sendJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", token, map[string]any{
		"match_date":    "2026-07-14",
		"match_time":    "12:00:00",
		"home_team_id":  homeTeamID,
		"guest_team_id": homeTeamID,
	})
	if createResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid request status %d, got %d with body %s", http.StatusBadRequest, createResponse.Code, createResponse.Body.String())
	}
}
