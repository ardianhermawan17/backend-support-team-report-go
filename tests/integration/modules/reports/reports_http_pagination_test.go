package reports

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

type reportsListMeta struct {
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	TotalItems  int  `json:"total_items"`
	TotalPages  int  `json:"total_pages"`
	HasNextPage bool `json:"has_next_page"`
	HasPrevPage bool `json:"has_prev_page"`
}

type reportsListPayload struct {
	Items []struct {
		ID int64 `json:"id"`
	} `json:"items"`
	Meta reportsListMeta `json:"meta"`
}

func TestReportsListPagination(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newReportsRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000001501, 7200000001501, "reports-pagination-admin", "reports-pagination-company", "reports-pagination-password")
	homeTeamID := createTeamForReportTests(t, router, token, "Reports Pagination Home")
	guestTeamID := createTeamForReportTests(t, router, token, "Reports Pagination Guest")

	createdIDs := make([]int64, 0, 7)
	for i := 1; i <= 7; i++ {
		scheduleID := createScheduleForReportTests(t, router, token, homeTeamID, guestTeamID, fmt.Sprintf("2026-10-%02d", i), "15:30:00")
		response := sendJSONRequest(t, router, http.MethodPost, "/api/v1/reports", token, map[string]any{
			"match_schedule_id":           scheduleID,
			"final_score_home":            i % 3,
			"final_score_guest":           (i + 1) % 3,
			"most_scoring_goal_player_id": nil,
		})
		if response.Code != http.StatusCreated {
			t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
		}
		createdReport := decodeReportResponse(t, response)
		createdIDs = append(createdIDs, createdReport.ID)
	}

	firstPage := listReportsPage(t, router, token, 1, 2)
	assertReportIDs(t, firstPage.Items, createdIDs[0:2])
	assertReportsMeta(t, firstPage.Meta, 1, 2, 7, 4, true, false)

	middlePage := listReportsPage(t, router, token, 2, 2)
	assertReportIDs(t, middlePage.Items, createdIDs[2:4])
	assertReportsMeta(t, middlePage.Meta, 2, 2, 7, 4, true, true)

	lastPage := listReportsPage(t, router, token, 4, 2)
	assertReportIDs(t, lastPage.Items, createdIDs[6:7])
	assertReportsMeta(t, lastPage.Meta, 4, 2, 7, 4, false, true)

	beyondPage := listReportsPage(t, router, token, 5, 2)
	if len(beyondPage.Items) != 0 {
		t.Fatalf("expected empty items beyond last page, got %#v", beyondPage.Items)
	}
	assertReportsMeta(t, beyondPage.Meta, 5, 2, 7, 4, false, true)

	emptyToken := createAccountAndLogin(t, accountRepo, router, 7100000001502, 7200000001502, "reports-pagination-admin-empty", "reports-pagination-company-empty", "reports-pagination-password-empty")
	emptyPage := listReportsPage(t, router, emptyToken, 1, 2)
	if len(emptyPage.Items) != 0 {
		t.Fatalf("expected empty list for company with no reports, got %#v", emptyPage.Items)
	}
	assertReportsMeta(t, emptyPage.Meta, 1, 2, 0, 0, false, false)

	invalidPageResponse := sendRequest(t, router, http.MethodGet, "/api/v1/reports?page=0&limit=2", token, nil)
	if invalidPageResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid page status %d, got %d with body %s", http.StatusBadRequest, invalidPageResponse.Code, invalidPageResponse.Body.String())
	}

	invalidLimitResponse := sendRequest(t, router, http.MethodGet, "/api/v1/reports?page=1&limit=0", token, nil)
	if invalidLimitResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid limit status %d, got %d with body %s", http.StatusBadRequest, invalidLimitResponse.Code, invalidLimitResponse.Body.String())
	}

	maxLimitPage := listReportsPage(t, router, token, 1, 100)
	if maxLimitPage.Meta.Limit != 50 {
		t.Fatalf("expected effective max limit 50, got %d", maxLimitPage.Meta.Limit)
	}
	if len(maxLimitPage.Items) != 7 {
		t.Fatalf("expected 7 items with max limit enforcement, got %d", len(maxLimitPage.Items))
	}

	stablePageOne := listReportsPage(t, router, token, 1, 3)
	stablePageTwo := listReportsPage(t, router, token, 2, 3)
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

func listReportsPage(t *testing.T, router http.Handler, token string, page, limit int) reportsListPayload {
	t.Helper()
	response := sendRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/reports?page=%d&limit=%d", page, limit), token, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload reportsListPayload
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal reports pagination payload: %v", err)
	}

	return payload
}

func assertReportIDs(t *testing.T, items []struct {
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

func assertReportsMeta(t *testing.T, actual reportsListMeta, page, limit, totalItems, totalPages int, hasNext, hasPrev bool) {
	t.Helper()
	if actual.Page != page || actual.Limit != limit || actual.TotalItems != totalItems || actual.TotalPages != totalPages || actual.HasNextPage != hasNext || actual.HasPrevPage != hasPrev {
		t.Fatalf("unexpected meta: %#v", actual)
	}
}
