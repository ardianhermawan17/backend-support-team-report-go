package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"backend-sport-team-report-go/internal/modules/auth/domain/entities"
	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestAccountRepositoryCreateAndFindByUsername(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	repo := authpersistence.NewAccountRepository(conn)

	account := newAccount(7100000000001, 7200000000001, "admin-alpha", "Alpha FC")

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

	if stored.User.Email != account.User.Email {
		t.Fatalf("expected user email %q, got %q", account.User.Email, stored.User.Email)
	}

	assertAuditLogCount(t, env.DB, "users", 1)
	assertAuditLogCount(t, env.DB, "companies", 1)
}

func TestAccountRepositoryCreateRequiresEmail(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	repo := authpersistence.NewAccountRepository(conn)

	account := newAccount(7100000000031, 7200000000031, "admin-missing-email", "Missing Email FC")
	account.User.Email = ""

	if err := repo.Create(context.Background(), account); err == nil {
		t.Fatal("expected create account without email to fail")
	}
}

func TestAccountRepositoryFindByUsernameIgnoresSoftDeletedAccounts(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	repo := authpersistence.NewAccountRepository(conn)

	account := newAccount(7100000000011, 7200000000011, "admin-beta", "Beta FC")

	if err := repo.Create(context.Background(), account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	if _, err := env.DB.Exec(`UPDATE users SET deleted_at = NOW() WHERE id = $1`, account.User.ID); err != nil {
		t.Fatalf("soft delete user: %v", err)
	}

	_, err := repo.FindByUsername(context.Background(), account.User.Username)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows after soft delete, got %v", err)
	}

	assertAuditActionCount(t, env.DB, "users", "SOFT_DELETE", 1)
	assertAuditActionCount(t, env.DB, "companies", "SOFT_DELETE", 1)
}

func TestAccountRepositoryFindByUsernameTreatsInjectionPayloadAsData(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	repo := authpersistence.NewAccountRepository(conn)

	account := newAccount(7100000000012, 7200000000012, "admin-sqli", "SQLi FC")

	if err := repo.Create(context.Background(), account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	_, err := repo.FindByUsername(context.Background(), `admin-sqli' OR '1'='1`)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for injection payload, got %v", err)
	}

	stored, err := repo.FindByUsername(context.Background(), account.User.Username)
	if err != nil {
		t.Fatalf("find account after injection lookup: %v", err)
	}

	if stored.User.ID != account.User.ID {
		t.Fatalf("expected stored user id %d, got %d", account.User.ID, stored.User.ID)
	}
}

func TestAccountRepositoryCreateAllowsReuseAfterSoftDelete(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	repo := authpersistence.NewAccountRepository(conn)

	first := newAccount(7100000000021, 7200000000021, "admin-gamma", "Gamma FC")

	if err := repo.Create(context.Background(), first); err != nil {
		t.Fatalf("create first account: %v", err)
	}

	if _, err := env.DB.Exec(`UPDATE users SET deleted_at = NOW() WHERE id = $1`, first.User.ID); err != nil {
		t.Fatalf("soft delete first user: %v", err)
	}

	second := newAccount(7100000000022, 7200000000022, first.User.Username, first.Company.Name)

	if err := repo.Create(context.Background(), second); err != nil {
		t.Fatalf("create second account with reused identifiers: %v", err)
	}

	stored, err := repo.FindByUsername(context.Background(), second.User.Username)
	if err != nil {
		t.Fatalf("find reused account: %v", err)
	}

	if stored.User.ID != second.User.ID {
		t.Fatalf("expected active user id %d, got %d", second.User.ID, stored.User.ID)
	}

	if stored.Company.ID != second.Company.ID {
		t.Fatalf("expected active company id %d, got %d", second.Company.ID, stored.Company.ID)
	}

	assertAuditActionCount(t, env.DB, "users", "SOFT_DELETE", 1)
	assertAuditActionCount(t, env.DB, "companies", "SOFT_DELETE", 1)
	assertAuditLogCount(t, env.DB, "users", 3)
	assertAuditLogCount(t, env.DB, "companies", 3)
}

func TestAccountRepositoryFindByUsernameTreatsMaliciousInputAsData(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	repo := authpersistence.NewAccountRepository(conn)

	account := newAccount(7100000000041, 7200000000041, "admin-safe-query", "Safe Query FC")
	if err := repo.Create(context.Background(), account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	var beforeCount int
	if err := env.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&beforeCount); err != nil {
		t.Fatalf("count users before malicious lookup: %v", err)
	}

	_, err := repo.FindByUsername(context.Background(), `' OR 1=1 --`)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for malicious username lookup, got %v", err)
	}

	var afterCount int
	if err := env.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&afterCount); err != nil {
		t.Fatalf("count users after malicious lookup: %v", err)
	}

	if afterCount != beforeCount {
		t.Fatalf("expected user count to remain %d, got %d", beforeCount, afterCount)
	}
}

func newAccount(userID, companyID int64, username, companyName string) entities.CompanyAdminAccount {
	return entities.CompanyAdminAccount{
		User: entities.User{
			ID:           userID,
			Username:     username,
			Email:        username + "@example.test",
			PasswordHash: "hashed-password",
		},
		Company: entities.Company{
			ID:   companyID,
			Name: companyName,
		},
	}
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

func assertAuditActionCount(t *testing.T, db *sql.DB, tableName, action string, expected int) {
	t.Helper()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM logs WHERE table_name = $1 AND action = $2`, tableName, action).Scan(&count); err != nil {
		t.Fatalf("count audit logs for %s with action %s: %v", tableName, action, err)
	}

	if count != expected {
		t.Fatalf("expected %d audit logs for %s with action %s, got %d", expected, tableName, action, count)
	}
}
