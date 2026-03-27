package integration

import (
	"context"
	"database/sql"
	"testing"

	"backend-sport-team-report-go/internal/modules/auth/domain/entities"
	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestAccountRepositoryCreateAndFindByUsername(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	repo := authpersistence.NewAccountRepository(conn)

	account := entities.CompanyAdminAccount{
		User: entities.User{
			ID:           7100000000001,
			Username:     "admin-alpha",
			PasswordHash: "hashed-password",
		},
		Company: entities.Company{
			ID:   7200000000001,
			Name: "Alpha FC",
		},
	}

	if err := repo.Create(context.Background(), account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	stored, err := repo.FindByUsername(context.Background(), account.User.Username)
	if err != nil {
		t.Fatalf("find account: %v", err)
	}

	if stored.User.ID != account.User.ID {
		t.Fatalf("expected user id %d, got %d", account.User.ID, stored.User.ID)
	}

	if stored.Company.ID != account.Company.ID {
		t.Fatalf("expected company id %d, got %d", account.Company.ID, stored.Company.ID)
	}

	if stored.Company.UserID != account.User.ID {
		t.Fatalf("expected company user id %d, got %d", account.User.ID, stored.Company.UserID)
	}

	assertAuditLogCount(t, env.DB, "users", 1)
	assertAuditLogCount(t, env.DB, "companies", 1)
}

func assertAuditLogCount(t *testing.T, db *sql.DB, tableName string, expected int) {
	t.Helper()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM logs WHERE table_name = $1`, tableName).Scan(&count); err != nil {
		t.Fatalf("count audit logs for %s: %v", tableName, err)
	}

	if count != expected {
		t.Fatalf("expected %d audit logs for %s, got %d", expected, tableName, count)
	}
}
