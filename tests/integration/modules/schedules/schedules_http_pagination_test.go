package schedules

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

type schedulesListMeta struct {
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	TotalItems  int  `json:"total_items"`
	TotalPages  int  `json:"total_pages"`
	HasNextPage bool `json:"has_next_page"`
	HasPrevPage bool `json:"has_prev_page"`
}

type schedulesListPayload struct {
	Items []struct {
		ID int64 `json:"id"`
	} `json:"items"`
	Meta schedulesListMeta `json:"meta"`
}

func TestSchedulesListPagination(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newSchedulesRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000001401, 7200000001401, "schedules-pagination-admin", "schedules-pagination-company", "schedules-pagination-password")
	homeTeamID := createTeamForScheduleTests(t, router, token, "Schedules Pagination Home")
	guestTeamID := createTeamForScheduleTests(t, router, token, "Schedules Pagination Guest")

	createdIDs := make([]int64, 0, 7)
	for i := 1; i <= 7; i++ {
		response := sendJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", token, map[string]any{
			"match_date":    fmt.Sprintf("2026-09-%02d", i),
			"match_time":    "15:30:00",
			"home_team_id":  homeTeamID,
			"guest_team_id": guestTeamID,
		})
		if response.Code != http.StatusCreated {
			t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
		}
		createdSchedule := decodeScheduleResponse(t, response)
		createdIDs = append(createdIDs, createdSchedule.ID)
	}

	firstPage := listSchedulesPage(t, router, token, 1, 2)
	assertScheduleIDs(t, firstPage.Items, createdIDs[0:2])
	assertSchedulesMeta(t, firstPage.Meta, 1, 2, 7, 4, true, false)

	middlePage := listSchedulesPage(t, router, token, 2, 2)
	assertScheduleIDs(t, middlePage.Items, createdIDs[2:4])
	assertSchedulesMeta(t, middlePage.Meta, 2, 2, 7, 4, true, true)

	lastPage := listSchedulesPage(t, router, token, 4, 2)
	assertScheduleIDs(t, lastPage.Items, createdIDs[6:7])
	assertSchedulesMeta(t, lastPage.Meta, 4, 2, 7, 4, false, true)

	beyondPage := listSchedulesPage(t, router, token, 5, 2)
	if len(beyondPage.Items) != 0 {
		t.Fatalf("expected empty items beyond last page, got %#v", beyondPage.Items)
	}
	assertSchedulesMeta(t, beyondPage.Meta, 5, 2, 7, 4, false, true)

	emptyToken := createAccountAndLogin(t, accountRepo, router, 7100000001402, 7200000001402, "schedules-pagination-admin-empty", "schedules-pagination-company-empty", "schedules-pagination-password-empty")
	emptyPage := listSchedulesPage(t, router, emptyToken, 1, 2)
	if len(emptyPage.Items) != 0 {
		t.Fatalf("expected empty list for company with no schedules, got %#v", emptyPage.Items)
	}
	assertSchedulesMeta(t, emptyPage.Meta, 1, 2, 0, 0, false, false)

	invalidPageResponse := sendRequest(t, router, http.MethodGet, "/api/v1/schedules?page=0&limit=2", token, nil)
	if invalidPageResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid page status %d, got %d with body %s", http.StatusBadRequest, invalidPageResponse.Code, invalidPageResponse.Body.String())
	}

	invalidLimitResponse := sendRequest(t, router, http.MethodGet, "/api/v1/schedules?page=1&limit=0", token, nil)
	if invalidLimitResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid limit status %d, got %d with body %s", http.StatusBadRequest, invalidLimitResponse.Code, invalidLimitResponse.Body.String())
	}

	maxLimitPage := listSchedulesPage(t, router, token, 1, 100)
	if maxLimitPage.Meta.Limit != 50 {
		t.Fatalf("expected effective max limit 50, got %d", maxLimitPage.Meta.Limit)
	}
	if len(maxLimitPage.Items) != 7 {
		t.Fatalf("expected 7 items with max limit enforcement, got %d", len(maxLimitPage.Items))
	}

	stablePageOne := listSchedulesPage(t, router, token, 1, 3)
	stablePageTwo := listSchedulesPage(t, router, token, 2, 3)
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

func listSchedulesPage(t *testing.T, router http.Handler, token string, page, limit int) schedulesListPayload {
	t.Helper()
	response := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/schedules?page=%d&limit=%d", page, limit), token, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload schedulesListPayload
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal schedules pagination payload: %v", err)
	}

	return payload
}

func assertScheduleIDs(t *testing.T, items []struct {
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

func assertSchedulesMeta(t *testing.T, actual schedulesListMeta, page, limit, totalItems, totalPages int, hasNext, hasPrev bool) {
	t.Helper()
	if actual.Page != page || actual.Limit != limit || actual.TotalItems != totalItems || actual.TotalPages != totalPages || actual.HasNextPage != hasNext || actual.HasPrevPage != hasPrev {
		t.Fatalf("unexpected meta: %#v", actual)
	}
}
