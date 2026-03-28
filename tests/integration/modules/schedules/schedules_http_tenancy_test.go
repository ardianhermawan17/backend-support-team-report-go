package schedules

import (
	"fmt"
	"net/http"
	"testing"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestSchedulesEnforcesCompanyBoundary(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newSchedulesRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	tokenCompanyA := createAccountAndLogin(t, accountRepo, router, 7100000000402, 7200000000402, "schedules-admin-b", "schedules-company-b", "schedules-password-b")
	tokenCompanyB := createAccountAndLogin(t, accountRepo, router, 7100000000403, 7200000000403, "schedules-admin-c", "schedules-company-c", "schedules-password-c")

	homeTeamID := createTeamForScheduleTests(t, router, tokenCompanyA, "Schedules Boundary Home FC")
	guestTeamID := createTeamForScheduleTests(t, router, tokenCompanyA, "Schedules Boundary Guest FC")

	createResponse := sendJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", tokenCompanyA, map[string]any{
		"match_date":    "2026-07-12",
		"match_time":    "14:15:00",
		"home_team_id":  homeTeamID,
		"guest_team_id": guestTeamID,
	})
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	createdSchedule := decodeScheduleResponse(t, createResponse)

	crossGet := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/schedules/%d", createdSchedule.ID), tokenCompanyB, nil)
	if crossGet.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company get status %d, got %d with body %s", http.StatusNotFound, crossGet.Code, crossGet.Body.String())
	}

	crossDelete := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/schedules/%d", createdSchedule.ID), tokenCompanyB, nil)
	if crossDelete.Code != http.StatusNotFound {
		t.Fatalf("expected cross-company delete status %d, got %d with body %s", http.StatusNotFound, crossDelete.Code, crossDelete.Body.String())
	}
}
