package seeding

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	ginrouter "backend-sport-team-report-go/internal/api/gin/router"
	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	appseeding "backend-sport-team-report-go/internal/platform/database/seeding"
	"backend-sport-team-report-go/internal/shared/logger"
	"backend-sport-team-report-go/tests/integration/testenv"
	"backend-sport-team-report-go/tests/integration/testhelpers"
)

const snowflakeNodeMask int64 = 1023

func TestSeedingSeedsBootstrapDataAndIsRepeatable(t *testing.T) {
	env := testenv.StartPostgres(t)
	service := newSeedService(t, env.DB)

	if err := service.Seed(context.Background()); err != nil {
		t.Fatalf("first seeding run: %v", err)
	}
	if err := service.Seed(context.Background()); err != nil {
		t.Fatalf("second seeding run: %v", err)
	}

	conn := env.OpenConnection(t)
	router := newRouter(t, conn)

	loginResponse := testhelpers.SendJSONRequest(t, router, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"username": "admin",
		"password": "password",
	})
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected seeded admin login status %d, got %d with body %s", http.StatusOK, loginResponse.Code, loginResponse.Body.String())
	}

	var loginPayload struct {
		AccessToken string `json:"access_token"`
		User        struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("unmarshal seeded login response: %v", err)
	}

	if loginPayload.User.Username != "admin" || loginPayload.User.Email != "admin@gmail.com" {
		t.Fatalf("unexpected seeded admin response: %#v", loginPayload.User)
	}

	assertActiveCount(t, env.DB, "users", 1)
	assertActiveCount(t, env.DB, "companies", 1)
	assertActiveCount(t, env.DB, "teams", 2)
	assertActiveCount(t, env.DB, "players", 2)
	assertActiveCount(t, env.DB, "schedules", 1)
	assertActiveCount(t, env.DB, "reports", 1)

	adminUserID := lookupID(t, env.DB, `SELECT id FROM users WHERE username = 'admin' AND deleted_at IS NULL`)
	companyID := lookupID(t, env.DB, `SELECT id FROM companies WHERE user_id = $1 AND deleted_at IS NULL`, adminUserID)
	homeTeamID := lookupID(t, env.DB, `SELECT id FROM teams WHERE company_id = $1 AND name = 'Admin United' AND deleted_at IS NULL`, companyID)
	guestTeamID := lookupID(t, env.DB, `SELECT id FROM teams WHERE company_id = $1 AND name = 'Seeder City' AND deleted_at IS NULL`, companyID)
	homePlayerID := lookupID(t, env.DB, `SELECT id FROM players WHERE team_id = $1 AND player_number = 9 AND deleted_at IS NULL`, homeTeamID)
	scheduleID := lookupID(t, env.DB, `SELECT id FROM schedules WHERE company_id = $1 AND home_team_id = $2 AND guest_team_id = $3 AND deleted_at IS NULL`, companyID, homeTeamID, guestTeamID)
	reportID := lookupID(t, env.DB, `SELECT id FROM reports WHERE match_schedule_id = $1 AND deleted_at IS NULL`, scheduleID)

	assertSnowflakeNode(t, adminUserID, 5)
	assertSnowflakeNode(t, companyID, 6)
	assertSnowflakeNode(t, homeTeamID, 1)
	assertSnowflakeNode(t, guestTeamID, 1)
	assertSnowflakeNode(t, homePlayerID, 2)
	assertSnowflakeNode(t, scheduleID, 3)
	assertSnowflakeNode(t, reportID, 4)

	assertSeededReportGraph(t, env.DB, scheduleID, reportID, homeTeamID, guestTeamID, homePlayerID)
}

func newSeedService(t *testing.T, db *sql.DB) *appseeding.Service {
	t.Helper()

	service, err := appseeding.NewService(db)
	if err != nil {
		t.Fatalf("create seeding service: %v", err)
	}

	return service
}

func newRouter(t *testing.T, conn *postgres.Connection) http.Handler {
	t.Helper()

	cfg := config.Config{
		App: config.AppConfig{
			Name: "soccer-team-report",
			Env:  config.EnvTest,
		},
		Database: config.DatabaseConfig{},
		Auth: config.AuthConfig{
			JWTSecret:      "integration-test-secret",
			AccessTokenTTL: 15 * time.Minute,
		},
	}

	return ginrouter.New(cfg, conn, logger.New(cfg.App.Name, cfg.App.Env))
}

func assertActiveCount(t *testing.T, db *sql.DB, tableName string, expected int) {
	t.Helper()

	query := `SELECT COUNT(*) FROM ` + tableName + ` WHERE deleted_at IS NULL`
	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		t.Fatalf("count active rows for %s: %v", tableName, err)
	}

	if count != expected {
		t.Fatalf("expected %d active rows in %s, got %d", expected, tableName, count)
	}
}

func lookupID(t *testing.T, db *sql.DB, query string, args ...any) int64 {
	t.Helper()

	var id int64
	if err := db.QueryRow(query, args...).Scan(&id); err != nil {
		t.Fatalf("lookup id with query %q: %v", query, err)
	}

	return id
}

func assertSnowflakeNode(t *testing.T, id int64, expectedNode int64) {
	t.Helper()

	if id <= 0 {
		t.Fatalf("expected positive snowflake id, got %d", id)
	}

	nodeID := (id >> 12) & snowflakeNodeMask
	if nodeID != expectedNode {
		t.Fatalf("expected snowflake node %d, got %d for id %d", expectedNode, nodeID, id)
	}
}

func assertSeededReportGraph(t *testing.T, db *sql.DB, scheduleID, reportID, homeTeamID, guestTeamID, topScorerID int64) {
	t.Helper()

	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM reports r
		JOIN schedules s ON s.id = r.match_schedule_id
		JOIN players p ON p.id = r.most_scoring_goal_player_id
		WHERE r.id = $1
		  AND s.id = $2
		  AND r.home_team_id = $3
		  AND r.guest_team_id = $4
		  AND p.id = $5
		  AND p.team_id = r.home_team_id
		  AND r.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND p.deleted_at IS NULL
	`, reportID, scheduleID, homeTeamID, guestTeamID, topScorerID).Scan(&count); err != nil {
		t.Fatalf("assert seeded report graph: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected seeded report graph to resolve exactly once, got %d", count)
	}
}
