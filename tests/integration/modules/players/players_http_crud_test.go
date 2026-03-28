package players

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestPlayersCRUDWithinTeamBoundary(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newPlayersRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000000301, 7200000000301, "players-admin-a", "players-company-a", "players-password-a")
	teamID := createTeamForPlayerTests(t, router, token, "Players Alpha FC")

	createResponse := sendJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/api/v1/teams/%d/players", teamID), token, map[string]any{
		"name":          "Rizky Pratama",
		"height":        180.5,
		"weight":        74.2,
		"position":      "striker",
		"player_number": 9,
	})
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	createdPlayer := decodePlayerResponse(t, createResponse)
	if createdPlayer.TeamID != teamID {
		t.Fatalf("expected team_id %d, got %d", teamID, createdPlayer.TeamID)
	}

	getResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d/players/%d", teamID, createdPlayer.ID), token, nil)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected get status %d, got %d with body %s", http.StatusOK, getResponse.Code, getResponse.Body.String())
	}

	listResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d/players", teamID), token, nil)
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
	if len(listPayload.Items) != 1 || listPayload.Items[0].ID != createdPlayer.ID {
		t.Fatalf("expected one player %d in list, got %#v", createdPlayer.ID, listPayload.Items)
	}

	updateResponse := sendJSONRequest(t, router, http.MethodPut, fmt.Sprintf("/api/v1/teams/%d/players/%d", teamID, createdPlayer.ID), token, map[string]any{
		"name":          "Rizky Pratama Updated",
		"height":        181.0,
		"weight":        75.0,
		"position":      "midfielder",
		"player_number": 8,
	})
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d with body %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}

	updatedPlayer := decodePlayerResponse(t, updateResponse)
	if updatedPlayer.Name != "Rizky Pratama Updated" {
		t.Fatalf("expected updated name, got %q", updatedPlayer.Name)
	}
	if updatedPlayer.Position != "midfielder" {
		t.Fatalf("expected updated position midfielder, got %q", updatedPlayer.Position)
	}

	deleteResponse := sendRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/teams/%d/players/%d", teamID, updatedPlayer.ID), token, nil)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected delete status %d, got %d with body %s", http.StatusNoContent, deleteResponse.Code, deleteResponse.Body.String())
	}

	getDeletedResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d/players/%d", teamID, updatedPlayer.ID), token, nil)
	if getDeletedResponse.Code != http.StatusNotFound {
		t.Fatalf("expected deleted get status %d, got %d with body %s", http.StatusNotFound, getDeletedResponse.Code, getDeletedResponse.Body.String())
	}

	var deletedAt *time.Time
	if err := env.DB.QueryRow(`SELECT deleted_at FROM players WHERE id = $1`, updatedPlayer.ID).Scan(&deletedAt); err != nil {
		t.Fatalf("query deleted_at: %v", err)
	}
	if deletedAt == nil {
		t.Fatal("expected deleted_at to be set after delete")
	}
}
