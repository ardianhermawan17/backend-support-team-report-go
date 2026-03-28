package teams

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestTeamsCRUDWithinCompany(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newTeamsRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000000201, 7200000000201, "teams-admin-a", "teams-company-a", "teams-password-a")

	createResponse := sendJSONRequest(t, router, http.MethodPost, "/api/v1/teams", token, map[string]any{
		"name":                     "Thunder FC",
		"founded_year":             2014,
		"homebase_address":         "Jalan Stadion Nomor 1",
		"city_of_homebase_address": "Bandung",
	})
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	createdTeam := decodeTeamResponse(t, createResponse)

	getResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d", createdTeam.ID), token, nil)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected get status %d, got %d with body %s", http.StatusOK, getResponse.Code, getResponse.Body.String())
	}

	listResponse := sendRequest(t, router, http.MethodGet, "/api/v1/teams", token, nil)
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
	if len(listPayload.Items) != 1 || listPayload.Items[0].ID != createdTeam.ID {
		t.Fatalf("expected one team %d in list, got %#v", createdTeam.ID, listPayload.Items)
	}

	updateResponse := sendJSONRequest(t, router, http.MethodPut, fmt.Sprintf("/api/v1/teams/%d", createdTeam.ID), token, map[string]any{
		"name":                     "Thunder FC Updated",
		"founded_year":             2015,
		"homebase_address":         "Jalan Stadion Nomor 99",
		"city_of_homebase_address": "Jakarta",
	})
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d with body %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}

	updatedTeam := decodeTeamResponse(t, updateResponse)
	if updatedTeam.Name != "Thunder FC Updated" {
		t.Fatalf("expected updated team name, got %q", updatedTeam.Name)
	}

	deleteResponse := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/teams/%d", updatedTeam.ID), token, nil)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected delete status %d, got %d with body %s", http.StatusNoContent, deleteResponse.Code, deleteResponse.Body.String())
	}

	getDeletedResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d", updatedTeam.ID), token, nil)
	if getDeletedResponse.Code != http.StatusNotFound {
		t.Fatalf("expected get deleted status %d, got %d with body %s", http.StatusNotFound, getDeletedResponse.Code, getDeletedResponse.Body.String())
	}

	var deletedAt *time.Time
	if err := env.DB.QueryRow(`SELECT deleted_at FROM teams WHERE id = $1`, updatedTeam.ID).Scan(&deletedAt); err != nil {
		t.Fatalf("query deleted_at: %v", err)
	}
	if deletedAt == nil {
		t.Fatal("expected deleted_at to be set after delete")
	}
}
