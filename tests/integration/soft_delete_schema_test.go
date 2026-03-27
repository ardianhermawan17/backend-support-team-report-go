package integration

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"backend-sport-team-report-go/internal/modules/auth/domain/entities"
	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestSoftDeleteAppliedToAllTables(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	repo := authpersistence.NewAccountRepository(conn)

	for _, tableName := range []string{"users", "companies", "images", "teams", "players", "schedules", "reports", "logs"} {
		assertTableHasDeletedAt(t, env.DB, tableName)
	}

	account := entities.CompanyAdminAccount{
		User: entities.User{
			ID:           8100000000001,
			Username:     "admin-soft-delete",
			PasswordHash: "hashed-password",
		},
		Company: entities.Company{
			ID:   8200000000001,
			Name: "Soft Delete FC",
		},
	}

	if err := repo.Create(context.Background(), account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	const (
		teamImageID   int64 = 8300000000001
		playerImageID int64 = 8300000000002
		homeTeamID    int64 = 8400000000001
		guestTeamID   int64 = 8400000000002
		playerID      int64 = 8500000000001
		scheduleID    int64 = 8600000000001
		reportID      int64 = 8700000000001
	)

	if _, err := env.DB.Exec(`
		INSERT INTO images (id, imageable_id, imageable_type, url, mime_type)
		VALUES
			($1, $2, 'team', 'https://example.com/team.png', 'image/png'),
			($3, $4, 'player', 'https://example.com/player.png', 'image/png')
	`, teamImageID, homeTeamID, playerImageID, playerID); err != nil {
		t.Fatalf("insert images: %v", err)
	}

	if _, err := env.DB.Exec(`
		INSERT INTO teams (id, company_id, name, logo_image_id, founded_year)
		VALUES
			($1, $2, 'Soft Delete Home', $3, 1999),
			($4, $2, 'Soft Delete Guest', NULL, 2001)
	`, homeTeamID, account.Company.ID, teamImageID, guestTeamID); err != nil {
		t.Fatalf("insert teams: %v", err)
	}

	if _, err := env.DB.Exec(`
		INSERT INTO players (id, team_id, name, position, player_number, profile_image_id)
		VALUES ($1, $2, 'Player One', 'striker', 9, $3)
	`, playerID, homeTeamID, playerImageID); err != nil {
		t.Fatalf("insert player: %v", err)
	}

	if _, err := env.DB.Exec(`
		INSERT INTO schedules (id, company_id, match_date, match_time, home_team_id, guest_team_id)
		VALUES ($1, $2, DATE '2026-03-27', TIME '15:00:00', $3, $4)
	`, scheduleID, account.Company.ID, homeTeamID, guestTeamID); err != nil {
		t.Fatalf("insert schedule: %v", err)
	}

	if _, err := env.DB.Exec(`
		INSERT INTO reports (
			id,
			match_schedule_id,
			home_team_id,
			guest_team_id,
			final_score_home,
			final_score_guest,
			status_match,
			most_scoring_goal_player_id
		) VALUES ($1, $2, $3, $4, 2, 1, 'home_team_win', $5)
	`, reportID, scheduleID, homeTeamID, guestTeamID, playerID); err != nil {
		t.Fatalf("insert report: %v", err)
	}

	if _, err := env.DB.Exec(`UPDATE users SET deleted_at = $1 WHERE id = $2`, time.Now().UTC(), account.User.ID); err != nil {
		t.Fatalf("soft delete user: %v", err)
	}

	assertRecordSoftDeleted(t, env.DB, "users", account.User.ID)
	assertRecordSoftDeleted(t, env.DB, "companies", account.Company.ID)
	assertRecordSoftDeleted(t, env.DB, "teams", homeTeamID)
	assertRecordSoftDeleted(t, env.DB, "teams", guestTeamID)
	assertRecordSoftDeleted(t, env.DB, "players", playerID)
	assertRecordSoftDeleted(t, env.DB, "schedules", scheduleID)
	assertRecordSoftDeleted(t, env.DB, "reports", reportID)
	assertRecordSoftDeleted(t, env.DB, "images", teamImageID)
	assertRecordSoftDeleted(t, env.DB, "images", playerImageID)

	assertSoftDeleteAuditCount(t, env.DB, "users", 1)
	assertSoftDeleteAuditCount(t, env.DB, "companies", 1)
	assertSoftDeleteAuditCount(t, env.DB, "teams", 2)
	assertSoftDeleteAuditCount(t, env.DB, "players", 1)
	assertSoftDeleteAuditCount(t, env.DB, "schedules", 1)
	assertSoftDeleteAuditCount(t, env.DB, "reports", 1)
	assertSoftDeleteAuditCount(t, env.DB, "images", 2)

	var logID int64
	if err := env.DB.QueryRow(`SELECT id FROM logs ORDER BY id LIMIT 1`).Scan(&logID); err != nil {
		t.Fatalf("select log row: %v", err)
	}

	if _, err := env.DB.Exec(`UPDATE logs SET deleted_at = $1 WHERE id = $2`, time.Now().UTC(), logID); err != nil {
		t.Fatalf("soft delete log row: %v", err)
	}

	assertRecordSoftDeleted(t, env.DB, "logs", logID)
}

func assertTableHasDeletedAt(t *testing.T, db *sql.DB, tableName string) {
	t.Helper()

	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name = $1
		  AND column_name = 'deleted_at'
	`, tableName).Scan(&count); err != nil {
		t.Fatalf("check deleted_at column for %s: %v", tableName, err)
	}

	if count != 1 {
		t.Fatalf("expected deleted_at column on %s", tableName)
	}
}

func assertRecordSoftDeleted(t *testing.T, db *sql.DB, tableName string, id int64) {
	t.Helper()

	query := `SELECT deleted_at IS NOT NULL FROM ` + tableName + ` WHERE id = $1`

	var softDeleted bool
	if err := db.QueryRow(query, id).Scan(&softDeleted); err != nil {
		t.Fatalf("check deleted row for %s(%d): %v", tableName, id, err)
	}

	if !softDeleted {
		t.Fatalf("expected %s(%d) to be soft deleted", tableName, id)
	}
}

func assertSoftDeleteAuditCount(t *testing.T, db *sql.DB, tableName string, expected int) {
	t.Helper()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM logs WHERE table_name = $1 AND action = 'SOFT_DELETE'`, tableName).Scan(&count); err != nil {
		t.Fatalf("count soft delete audit logs for %s: %v", tableName, err)
	}

	if count != expected {
		t.Fatalf("expected %d soft delete audit logs for %s, got %d", expected, tableName, count)
	}
}
