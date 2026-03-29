package teams

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

type teamsListMeta struct {
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	TotalItems  int  `json:"total_items"`
	TotalPages  int  `json:"total_pages"`
	HasNextPage bool `json:"has_next_page"`
	HasPrevPage bool `json:"has_prev_page"`
}

type teamsListPayload struct {
	Items []struct {
		ID int64 `json:"id"`
	} `json:"items"`
	Meta teamsListMeta `json:"meta"`
}

func TestTeamsListPagination(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newTeamsRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000001201, 7200000001201, "teams-pagination-admin", "teams-pagination-company", "teams-pagination-password")

	createdIDs := make([]int64, 0, 7)
	for i := 1; i <= 7; i++ {
		response := sendJSONRequest(t, router, http.MethodPost, "/api/v1/teams", token, map[string]any{
			"name":                     fmt.Sprintf("Pagination Team %02d", i),
			"founded_year":             2010 + i,
			"homebase_address":         "Jalan Stadion Nomor 1",
			"city_of_homebase_address": "Bandung",
		})
		if response.Code != http.StatusCreated {
			t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
		}
		createdTeam := decodeTeamResponse(t, response)
		createdIDs = append(createdIDs, createdTeam.ID)
	}

	firstPage := listTeamsPage(t, router, token, 1, 2)
	assertTeamIDs(t, firstPage.Items, createdIDs[0:2])
	assertTeamsMeta(t, firstPage.Meta, 1, 2, 7, 4, true, false)

	middlePage := listTeamsPage(t, router, token, 2, 2)
	assertTeamIDs(t, middlePage.Items, createdIDs[2:4])
	assertTeamsMeta(t, middlePage.Meta, 2, 2, 7, 4, true, true)

	lastPage := listTeamsPage(t, router, token, 4, 2)
	assertTeamIDs(t, lastPage.Items, createdIDs[6:7])
	assertTeamsMeta(t, lastPage.Meta, 4, 2, 7, 4, false, true)

	beyondPage := listTeamsPage(t, router, token, 5, 2)
	if len(beyondPage.Items) != 0 {
		t.Fatalf("expected empty items beyond last page, got %#v", beyondPage.Items)
	}
	assertTeamsMeta(t, beyondPage.Meta, 5, 2, 7, 4, false, true)

	emptyToken := createAccountAndLogin(t, accountRepo, router, 7100000001202, 7200000001202, "teams-pagination-admin-empty", "teams-pagination-company-empty", "teams-pagination-password-empty")
	emptyPage := listTeamsPage(t, router, emptyToken, 1, 2)
	if len(emptyPage.Items) != 0 {
		t.Fatalf("expected empty list for new company, got %#v", emptyPage.Items)
	}
	assertTeamsMeta(t, emptyPage.Meta, 1, 2, 0, 0, false, false)

	invalidPageResponse := sendRequest(t, router, http.MethodGet, "/api/v1/teams?page=0&limit=2", token, nil)
	if invalidPageResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid page status %d, got %d with body %s", http.StatusBadRequest, invalidPageResponse.Code, invalidPageResponse.Body.String())
	}

	invalidLimitResponse := sendRequest(t, router, http.MethodGet, "/api/v1/teams?page=1&limit=0", token, nil)
	if invalidLimitResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid limit status %d, got %d with body %s", http.StatusBadRequest, invalidLimitResponse.Code, invalidLimitResponse.Body.String())
	}

	maxLimitResponse := listTeamsPage(t, router, token, 1, 100)
	if maxLimitResponse.Meta.Limit != 50 {
		t.Fatalf("expected effective max limit 50, got %d", maxLimitResponse.Meta.Limit)
	}
	if len(maxLimitResponse.Items) != 7 {
		t.Fatalf("expected 7 items with max limit enforcement, got %d", len(maxLimitResponse.Items))
	}

	stablePageOne := listTeamsPage(t, router, token, 1, 3)
	stablePageTwo := listTeamsPage(t, router, token, 2, 3)
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

func listTeamsPage(t *testing.T, router http.Handler, token string, page, limit int) teamsListPayload {
	t.Helper()
	response := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/teams?page=%d&limit=%d", page, limit), token, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload teamsListPayload
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal teams pagination payload: %v", err)
	}

	return payload
}

func assertTeamIDs(t *testing.T, items []struct {
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

func assertTeamsMeta(t *testing.T, actual teamsListMeta, page, limit, totalItems, totalPages int, hasNext, hasPrev bool) {
	t.Helper()
	if actual.Page != page || actual.Limit != limit || actual.TotalItems != totalItems || actual.TotalPages != totalPages || actual.HasNextPage != hasNext || actual.HasPrevPage != hasPrev {
		t.Fatalf("unexpected meta: %#v", actual)
	}
}
