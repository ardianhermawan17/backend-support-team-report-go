package schedules

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestSchedulesCRUDWithinCompany(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newSchedulesRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000000401, 7200000000401, "schedules-admin-a", "schedules-company-a", "schedules-password-a")
	homeTeamID := createTeamForScheduleTests(t, router, token, "Schedules Home FC")
	guestTeamID := createTeamForScheduleTests(t, router, token, "Schedules Guest FC")

	createResponse := sendJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", token, map[string]any{
		"match_date":    "2026-07-10",
		"match_time":    "15:30:00",
		"home_team_id":  homeTeamID,
		"guest_team_id": guestTeamID,
	})
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	createdSchedule := decodeScheduleResponse(t, createResponse)

	getResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/schedules/%d", createdSchedule.ID), token, nil)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected get status %d, got %d with body %s", http.StatusOK, getResponse.Code, getResponse.Body.String())
	}

	listResponse := sendRequest(t, router, http.MethodGet, "/api/v1/schedules", token, nil)
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
	if len(listPayload.Items) != 1 || listPayload.Items[0].ID != createdSchedule.ID {
		t.Fatalf("expected one schedule %d in list, got %#v", createdSchedule.ID, listPayload.Items)
	}

	updateResponse := sendJSONRequest(t, router, http.MethodPut, fmt.Sprintf("/api/v1/schedules/%d", createdSchedule.ID), token, map[string]any{
		"match_date":    "2026-07-11",
		"match_time":    "16:00:00",
		"home_team_id":  homeTeamID,
		"guest_team_id": guestTeamID,
	})
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d with body %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}

	updatedSchedule := decodeScheduleResponse(t, updateResponse)
	if updatedSchedule.MatchDate != "2026-07-11" {
		t.Fatalf("expected updated schedule date, got %q", updatedSchedule.MatchDate)
	}

	deleteResponse := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/schedules/%d", updatedSchedule.ID), token, nil)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected delete status %d, got %d with body %s", http.StatusNoContent, deleteResponse.Code, deleteResponse.Body.String())
	}

	getDeletedResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/schedules/%d", updatedSchedule.ID), token, nil)
	if getDeletedResponse.Code != http.StatusNotFound {
		t.Fatalf("expected get deleted status %d, got %d with body %s", http.StatusNotFound, getDeletedResponse.Code, getDeletedResponse.Body.String())
	}

	var deletedAt *time.Time
	if err := env.DB.QueryRow(`SELECT deleted_at FROM schedules WHERE id = $1`, updatedSchedule.ID).Scan(&deletedAt); err != nil {
		t.Fatalf("query deleted_at: %v", err)
	}
	if deletedAt == nil {
		t.Fatal("expected deleted_at to be set after delete")
	}
}
