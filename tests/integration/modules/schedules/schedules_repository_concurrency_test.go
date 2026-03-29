package schedules

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	scheduledomain "backend-sport-team-report-go/internal/modules/schedule/domain"
	schedulespersistence "backend-sport-team-report-go/internal/modules/schedule/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestScheduleRepositoryRejectsStaleConcurrentUpdates(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newSchedulesRouter(conn)
	accountRepo := authpersistence.NewAccountRepository(conn)
	scheduleRepo := schedulespersistence.NewScheduleRepository(conn)

	const userID int64 = 7100000000412
	const companyID int64 = 7200000000412
	token := createAccountAndLogin(t, accountRepo, router, userID, companyID, "schedules-admin-stale", "schedules-company-stale", "schedules-password-stale")
	homeTeamID := createTeamForScheduleTests(t, router, token, "Schedules Stale Home FC")
	guestTeamID := createTeamForScheduleTests(t, router, token, "Schedules Stale Guest FC")
	scheduleID := createScheduleForConcurrencyTests(t, router, token, homeTeamID, guestTeamID, "2026-08-23", "10:00:00")

	baseSchedule, err := scheduleRepo.FindByIDAndCompany(context.Background(), scheduleID, companyID)
	if err != nil {
		t.Fatalf("find base schedule: %v", err)
	}

	firstUpdate := baseSchedule
	firstUpdate.MatchTime = mustParseMatchTime(t, "17:00:00")

	secondUpdate := baseSchedule
	secondUpdate.MatchTime = mustParseMatchTime(t, "18:00:00")

	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup

	for _, scheduledUpdate := range []struct {
		scheduleTime time.Time
	}{
		{scheduleTime: firstUpdate.MatchTime},
		{scheduleTime: secondUpdate.MatchTime},
	} {
		wg.Add(1)
		go func(updateTime time.Time) {
			defer wg.Done()
			<-start

			staleCopy := baseSchedule
			staleCopy.MatchTime = updateTime
			results <- scheduleRepo.Update(context.Background(), staleCopy, userID)
		}(scheduledUpdate.scheduleTime)
	}

	close(start)
	wg.Wait()
	close(results)

	var successCount int
	var conflictCount int
	for updateErr := range results {
		switch {
		case updateErr == nil:
			successCount++
		case errors.Is(updateErr, scheduledomain.ErrScheduleConcurrentModification):
			conflictCount++
		default:
			t.Fatalf("expected success or concurrent modification, got %v", updateErr)
		}
	}

	if successCount != 1 || conflictCount != 1 {
		t.Fatalf("expected one successful update and one concurrent modification, got success=%d conflict=%d", successCount, conflictCount)
	}

	stored, err := scheduleRepo.FindByIDAndCompany(context.Background(), scheduleID, companyID)
	if err != nil {
		t.Fatalf("find updated schedule: %v", err)
	}

	storedTime := stored.MatchTime.Format("15:04:05")
	if storedTime != "17:00:00" && storedTime != "18:00:00" {
		t.Fatalf("expected persisted match_time to be one of the concurrent updates, got %s", storedTime)
	}
}

func createScheduleForConcurrencyTests(t *testing.T, router http.Handler, token string, homeTeamID, guestTeamID int64, matchDate, matchTime string) int64 {
	t.Helper()

	response := sendJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", token, map[string]any{
		"match_date":    matchDate,
		"match_time":    matchTime,
		"home_team_id":  homeTeamID,
		"guest_team_id": guestTeamID,
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("expected create schedule status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal create schedule response: %v", err)
	}

	return payload.ID
}

func mustParseMatchTime(t *testing.T, value string) time.Time {
	t.Helper()

	matchTime, err := time.Parse("15:04:05", value)
	if err != nil {
		t.Fatalf("parse match time %s: %v", value, err)
	}

	return matchTime
}
