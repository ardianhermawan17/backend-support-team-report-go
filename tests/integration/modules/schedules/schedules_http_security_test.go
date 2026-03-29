package schedules

import (
	"net/http"
	"sync"
	"testing"
	"time"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
	"backend-sport-team-report-go/tests/integration/testhelpers"
)

func TestSchedulesCreateRateLimitRejectsRepeatedRequests(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	security := testhelpers.DefaultTestConfig().Security
	security.RateLimit.AuthenticatedWrite.Window = time.Minute
	security.RateLimit.AuthenticatedWrite.MaxRequests = 2
	router := newSchedulesRouterWithSecurity(conn, security)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000000410, 7200000000410, "schedules-admin-rate", "schedules-company-rate", "schedules-password-rate")
	homeTeamID := createTeamForScheduleTests(t, router, token, "Schedules Rate Home FC")
	guestTeamID := createTeamForScheduleTests(t, router, token, "Schedules Rate Guest FC")

	for attempt := 1; attempt <= 3; attempt++ {
		response := sendJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", token, map[string]any{
			"match_date":    "2026-08-21",
			"match_time":    []string{"11:00:00", "12:00:00", "13:00:00"}[attempt-1],
			"home_team_id":  homeTeamID,
			"guest_team_id": guestTeamID,
		})

		switch attempt {
		case 1, 2:
			if response.Code != http.StatusCreated {
				t.Fatalf("expected created status %d on attempt %d, got %d with body %s", http.StatusCreated, attempt, response.Code, response.Body.String())
			}
		case 3:
			if response.Code != http.StatusTooManyRequests {
				t.Fatalf("expected too many requests status %d on attempt %d, got %d with body %s", http.StatusTooManyRequests, attempt, response.Code, response.Body.String())
			}
			if response.Header().Get("Retry-After") == "" {
				t.Fatal("expected Retry-After header on throttled schedule create response")
			}
		}
	}
}

func TestSchedulesConcurrentDuplicateCreateReturnsConflict(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newSchedulesRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)

	token := createAccountAndLogin(t, accountRepo, router, 7100000000411, 7200000000411, "schedules-admin-concurrent", "schedules-company-concurrent", "schedules-password-concurrent")
	homeTeamID := createTeamForScheduleTests(t, router, token, "Schedules Concurrent Home FC")
	guestTeamID := createTeamForScheduleTests(t, router, token, "Schedules Concurrent Guest FC")

	start := make(chan struct{})
	results := make(chan int, 2)
	var wg sync.WaitGroup

	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			response := sendJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", token, map[string]any{
				"match_date":    "2026-08-22",
				"match_time":    "15:00:00",
				"home_team_id":  homeTeamID,
				"guest_team_id": guestTeamID,
			})
			results <- response.Code
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	var createdCount int
	var conflictCount int
	for statusCode := range results {
		switch statusCode {
		case http.StatusCreated:
			createdCount++
		case http.StatusConflict:
			conflictCount++
		default:
			t.Fatalf("expected only created/conflict responses, got %d", statusCode)
		}
	}

	if createdCount != 1 || conflictCount != 1 {
		t.Fatalf("expected one created and one conflict response, got created=%d conflict=%d", createdCount, conflictCount)
	}
}
