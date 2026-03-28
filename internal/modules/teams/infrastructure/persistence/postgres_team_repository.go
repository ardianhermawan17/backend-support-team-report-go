package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"

	teamdomain "backend-sport-team-report-go/internal/modules/teams/domain"
	"backend-sport-team-report-go/internal/modules/teams/domain/entities"
	"backend-sport-team-report-go/internal/platform/database/postgres"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(conn *postgres.Connection) *TeamRepository {
	return &TeamRepository{db: conn.DB()}
}

func (r *TeamRepository) Create(ctx context.Context, team entities.Team, actorUserID int64) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin team create transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, strconv.FormatInt(actorUserID, 10)); err != nil {
		return fmt.Errorf("set audit actor for team create: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO teams (id, company_id, name, logo_image_id, founded_year, homebase_address, city_of_homebase_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, team.ID, team.CompanyID, team.Name, team.LogoImageID, team.FoundedYear, team.HomebaseAddress, team.CityOfHomebaseAddress); err != nil {
		return fmt.Errorf("insert team: %w", classifyWriteError(err))
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit team create transaction: %w", err)
	}

	return nil
}

func (r *TeamRepository) ListByCompany(ctx context.Context, companyID int64) ([]entities.Team, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			company_id,
			name,
			logo_image_id,
			founded_year,
			homebase_address,
			city_of_homebase_address,
			created_at,
			updated_at,
			deleted_at
		FROM teams
		WHERE company_id = $1
		  AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, companyID)
	if err != nil {
		return nil, fmt.Errorf("list teams by company: %w", err)
	}
	defer rows.Close()

	teams := make([]entities.Team, 0)
	for rows.Next() {
		team, err := scanTeam(rows)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate teams by company: %w", err)
	}

	return teams, nil
}

func (r *TeamRepository) FindByIDAndCompany(ctx context.Context, teamID, companyID int64) (entities.Team, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			company_id,
			name,
			logo_image_id,
			founded_year,
			homebase_address,
			city_of_homebase_address,
			created_at,
			updated_at,
			deleted_at
		FROM teams
		WHERE id = $1
		  AND company_id = $2
		  AND deleted_at IS NULL
	`, teamID, companyID)

	team, err := scanTeam(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.Team{}, teamdomain.ErrTeamNotFound
		}
		return entities.Team{}, fmt.Errorf("find team by id and company: %w", err)
	}

	return team, nil
}

func (r *TeamRepository) Update(ctx context.Context, team entities.Team, actorUserID int64) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin team update transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, strconv.FormatInt(actorUserID, 10)); err != nil {
		return fmt.Errorf("set audit actor for team update: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE teams
		SET
			name = $1,
			logo_image_id = $2,
			founded_year = $3,
			homebase_address = $4,
			city_of_homebase_address = $5
		WHERE id = $6
		  AND company_id = $7
		  AND deleted_at IS NULL
	`, team.Name, team.LogoImageID, team.FoundedYear, team.HomebaseAddress, team.CityOfHomebaseAddress, team.ID, team.CompanyID)
	if err != nil {
		return fmt.Errorf("update team: %w", classifyWriteError(err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read team update affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return teamdomain.ErrTeamNotFound
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit team update transaction: %w", err)
	}

	return nil
}

func (r *TeamRepository) SoftDelete(ctx context.Context, teamID, companyID, actorUserID int64) (deleted bool, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin team delete transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, strconv.FormatInt(actorUserID, 10)); err != nil {
		return false, fmt.Errorf("set audit actor for team delete: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE teams
		SET deleted_at = NOW()
		WHERE id = $1
		  AND company_id = $2
		  AND deleted_at IS NULL
	`, teamID, companyID)
	if err != nil {
		return false, fmt.Errorf("soft delete team: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read team delete affected rows: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("commit team delete transaction: %w", err)
	}

	return rowsAffected > 0, nil
}

type teamScanner interface {
	Scan(dest ...any) error
}

func scanTeam(scanner teamScanner) (entities.Team, error) {
	var team entities.Team
	var logoImageID sql.NullInt64
	var deletedAt sql.NullTime

	err := scanner.Scan(
		&team.ID,
		&team.CompanyID,
		&team.Name,
		&logoImageID,
		&team.FoundedYear,
		&team.HomebaseAddress,
		&team.CityOfHomebaseAddress,
		&team.CreatedAt,
		&team.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		return entities.Team{}, err
	}

	if logoImageID.Valid {
		logoID := logoImageID.Int64
		team.LogoImageID = &logoID
	}
	if deletedAt.Valid {
		deletedTimestamp := deletedAt.Time
		team.DeletedAt = &deletedTimestamp
	}

	return team, nil
}

func classifyWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "uq_teams_company_name_active":
				return teamdomain.ErrTeamAlreadyExists
			case "uq_teams_logo_image_id_active":
				return teamdomain.ErrTeamLogoAlreadyInUse
			}
		}
	}

	return err
}
