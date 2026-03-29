package players

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

type playersListMeta struct {
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	TotalItems  int  `json:"total_items"`
	TotalPages  int  `json:"total_pages"`
	HasNextPage bool `json:"has_next_page"`
	HasPrevPage bool `json:"has_prev_page"`
}

type playersListPayload struct {
	Items []struct {
		ID int64 `json:"id"`
	} `json:"items"`
	Meta playersListMeta `json:"meta"`
}

func TestPlayersListPagination(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newPlayersRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000001301, 7200000001301, "players-pagination-admin", "players-pagination-company", "players-pagination-password")
	teamID := createTeamForPlayerTests(t, router, token, "Players Pagination Team")

	createdIDs := make([]int64, 0, 7)
	for i := 1; i <= 7; i++ {
		response := sendJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/api/v1/teams/%d/players", teamID), token, map[string]any{
			"name":          fmt.Sprintf("Player Pagination %02d", i),
			"height":        170.0 + float64(i),
			"weight":        65.0 + float64(i),
			"position":      "striker",
			"player_number": i,
		})
		if response.Code != http.StatusCreated {
			t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
		}
		createdPlayer := decodePlayerResponse(t, response)
		createdIDs = append(createdIDs, createdPlayer.ID)
	}

	firstPage := listPlayersPage(t, router, token, teamID, 1, 2)
	assertPlayerIDs(t, firstPage.Items, createdIDs[0:2])
	assertPlayersMeta(t, firstPage.Meta, 1, 2, 7, 4, true, false)

	middlePage := listPlayersPage(t, router, token, teamID, 2, 2)
	assertPlayerIDs(t, middlePage.Items, createdIDs[2:4])
	assertPlayersMeta(t, middlePage.Meta, 2, 2, 7, 4, true, true)

	lastPage := listPlayersPage(t, router, token, teamID, 4, 2)
	assertPlayerIDs(t, lastPage.Items, createdIDs[6:7])
	assertPlayersMeta(t, lastPage.Meta, 4, 2, 7, 4, false, true)

	beyondPage := listPlayersPage(t, router, token, teamID, 5, 2)
	if len(beyondPage.Items) != 0 {
		t.Fatalf("expected empty items beyond last page, got %#v", beyondPage.Items)
	}
	assertPlayersMeta(t, beyondPage.Meta, 5, 2, 7, 4, false, true)

	emptyTeamID := createTeamForPlayerTests(t, router, token, "Players Empty Team")
	emptyPage := listPlayersPage(t, router, token, emptyTeamID, 1, 2)
	if len(emptyPage.Items) != 0 {
		t.Fatalf("expected empty list for team with no players, got %#v", emptyPage.Items)
	}
	assertPlayersMeta(t, emptyPage.Meta, 1, 2, 0, 0, false, false)

	invalidPageResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d/players?page=0&limit=2", teamID), token, nil)
	if invalidPageResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid page status %d, got %d with body %s", http.StatusBadRequest, invalidPageResponse.Code, invalidPageResponse.Body.String())
	}

	invalidLimitResponse := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d/players?page=1&limit=0", teamID), token, nil)
	if invalidLimitResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid limit status %d, got %d with body %s", http.StatusBadRequest, invalidLimitResponse.Code, invalidLimitResponse.Body.String())
	}

	maxLimitPage := listPlayersPage(t, router, token, teamID, 1, 100)
	if maxLimitPage.Meta.Limit != 50 {
		t.Fatalf("expected effective max limit 50, got %d", maxLimitPage.Meta.Limit)
	}
	if len(maxLimitPage.Items) != 7 {
		t.Fatalf("expected 7 items with max limit enforcement, got %d", len(maxLimitPage.Items))
	}

	stablePageOne := listPlayersPage(t, router, token, teamID, 1, 3)
	stablePageTwo := listPlayersPage(t, router, token, teamID, 2, 3)
	combined := []int64{
		stablePageOne.Items[0].ID,
		stablePageOne.Items[1].ID,
		stablePageOne.Items[2].ID,
		stablePageTwo.Items[0].ID,
		stablePageTwo.Items[1].ID,
		stablePageTwo.Items[2].ID,
	}
	for i := 0; i < len(combined); i++ {
		if combined[i] != createdIDs[i] {
			t.Fatalf("expected stable ordering id %d at index %d, got %d", createdIDs[i], i, combined[i])
		}
	}
}

func listPlayersPage(t *testing.T, router http.Handler, token string, teamID int64, page, limit int) playersListPayload {
	t.Helper()
	response := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams/%d/players?page=%d&limit=%d", teamID, page, limit), token, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload playersListPayload
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal players pagination payload: %v", err)
	}

	return payload
}

func assertPlayerIDs(t *testing.T, items []struct {
	ID int64 `json:"id"`
}, expected []int64) {
	t.Helper()
	if len(items) != len(expected) {
		t.Fatalf("expected %d items, got %d", len(expected), len(items))
	}
	for i := range expected {
		if items[i].ID != expected[i] {
			t.Fatalf("expected id %d at index %d, got %d", expected[i], i, items[i].ID)
		}
	}
}

func assertPlayersMeta(t *testing.T, actual playersListMeta, page, limit, totalItems, totalPages int, hasNext, hasPrev bool) {
	t.Helper()
	if actual.Page != page || actual.Limit != limit || actual.TotalItems != totalItems || actual.TotalPages != totalPages || actual.HasNextPage != hasNext || actual.HasPrevPage != hasPrev {
		t.Fatalf("unexpected meta: %#v", actual)
	}
}
