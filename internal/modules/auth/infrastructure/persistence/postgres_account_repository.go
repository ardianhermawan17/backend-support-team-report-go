package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"backend-sport-team-report-go/internal/modules/auth/domain/entities"
	"backend-sport-team-report-go/internal/platform/database/postgres"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(conn *postgres.Connection) *AccountRepository {
	return &AccountRepository{db: conn.DB()}
}

func (r *AccountRepository) Create(ctx context.Context, account entities.CompanyAdminAccount) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin account transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO users (id, username, password_hash)
		VALUES ($1, $2, $3)
	`, account.User.ID, account.User.Username, account.User.PasswordHash); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, fmt.Sprintf("%d", account.User.ID)); err != nil {
		return fmt.Errorf("set audit actor: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO companies (id, user_id, name)
		VALUES ($1, $2, $3)
	`, account.Company.ID, account.User.ID, account.Company.Name); err != nil {
		return fmt.Errorf("insert company: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit account transaction: %w", err)
	}

	return nil
}

func (r *AccountRepository) FindByUsername(ctx context.Context, username string) (entities.CompanyAdminAccount, error) {
	var account entities.CompanyAdminAccount

	err := r.db.QueryRowContext(ctx, `
		SELECT
			u.id,
			u.username,
			u.password_hash,
			u.created_at,
			u.updated_at,
			c.id,
			c.user_id,
			c.name,
			c.created_at,
			c.updated_at
		FROM users u
		JOIN companies c ON c.user_id = u.id
		WHERE u.username = $1
		  AND u.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		`, username).Scan(
		&account.User.ID,
		&account.User.Username,
		&account.User.PasswordHash,
		&account.User.CreatedAt,
		&account.User.UpdatedAt,
		&account.Company.ID,
		&account.Company.UserID,
		&account.Company.Name,
		&account.Company.CreatedAt,
		&account.Company.UpdatedAt,
	)
	if err != nil {
		return entities.CompanyAdminAccount{}, fmt.Errorf("find account by username: %w", err)
	}

	return account, nil
}
