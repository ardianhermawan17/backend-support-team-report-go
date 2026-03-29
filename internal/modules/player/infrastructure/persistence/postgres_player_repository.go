package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"

	playerdomain "backend-sport-team-report-go/internal/modules/player/domain"
	"backend-sport-team-report-go/internal/modules/player/domain/entities"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/paginator"
)

type PlayerRepository struct {
	db *sql.DB
}

func NewPlayerRepository(conn *postgres.Connection) *PlayerRepository {
	return &PlayerRepository{db: conn.DB()}
}

func (r *PlayerRepository) Create(ctx context.Context, companyID int64, player entities.Player, actorUserID int64) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin player create transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, strconv.FormatInt(actorUserID, 10)); err != nil {
		return fmt.Errorf("set audit actor for player create: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO players (id, team_id, name, height, weight, position, player_number, profile_image_id)
		SELECT $1, t.id, $2, $3, $4, $5, $6, $7
		FROM teams t
		WHERE t.id = $8
		  AND t.company_id = $9
		  AND t.deleted_at IS NULL
	`, player.ID, player.Name, player.Height, player.Weight, player.Position, player.PlayerNumber, player.ProfileImageID, player.TeamID, companyID)
	if err != nil {
		return fmt.Errorf("insert player: %w", classifyWriteError(err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read player create affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return playerdomain.ErrTeamNotFound
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit player create transaction: %w", err)
	}

	return nil
}

func (r *PlayerRepository) ListByTeam(ctx context.Context, companyID, teamID int64, params paginator.Params) (paginator.Result[entities.Player], error) {
	if !r.teamExists(ctx, companyID, teamID) {
		return paginator.Result[entities.Player]{}, playerdomain.ErrTeamNotFound
	}

	var totalItems int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM players p
		JOIN teams t ON t.id = p.team_id
		WHERE p.team_id = $1
		  AND t.company_id = $2
		  AND t.deleted_at IS NULL
		  AND p.deleted_at IS NULL
	`, teamID, companyID).Scan(&totalItems); err != nil {
		return paginator.Result[entities.Player]{}, fmt.Errorf("count players by team: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			p.id,
			p.team_id,
			p.name,
			p.height,
			p.weight,
			p.position,
			p.player_number,
			p.profile_image_id,
			p.created_at,
			p.updated_at,
			p.deleted_at
		FROM players p
		JOIN teams t ON t.id = p.team_id
		WHERE p.team_id = $1
		  AND t.company_id = $2
		  AND t.deleted_at IS NULL
		  AND p.deleted_at IS NULL
		ORDER BY p.created_at ASC, p.id ASC
		LIMIT $3
		OFFSET $4
	`, teamID, companyID, params.Limit, params.Offset)
	if err != nil {
		return paginator.Result[entities.Player]{}, fmt.Errorf("list players by team: %w", err)
	}
	defer rows.Close()

	players := make([]entities.Player, 0)
	for rows.Next() {
		player, scanErr := scanPlayer(rows)
		if scanErr != nil {
			return paginator.Result[entities.Player]{}, scanErr
		}
		players = append(players, player)
	}

	if err := rows.Err(); err != nil {
		return paginator.Result[entities.Player]{}, fmt.Errorf("iterate players by team: %w", err)
	}

	return paginator.Result[entities.Player]{
		Items: players,
		Meta:  paginator.BuildMeta(params, totalItems),
	}, nil
}

func (r *PlayerRepository) FindByID(ctx context.Context, companyID, teamID, playerID int64) (entities.Player, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			p.id,
			p.team_id,
			p.name,
			p.height,
			p.weight,
			p.position,
			p.player_number,
			p.profile_image_id,
			p.created_at,
			p.updated_at,
			p.deleted_at
		FROM players p
		JOIN teams t ON t.id = p.team_id
		WHERE p.id = $1
		  AND p.team_id = $2
		  AND t.company_id = $3
		  AND t.deleted_at IS NULL
		  AND p.deleted_at IS NULL
	`, playerID, teamID, companyID)

	player, err := scanPlayer(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if !r.teamExists(ctx, companyID, teamID) {
				return entities.Player{}, playerdomain.ErrTeamNotFound
			}
			return entities.Player{}, playerdomain.ErrPlayerNotFound
		}
		return entities.Player{}, fmt.Errorf("find player by id: %w", err)
	}

	return player, nil
}

func (r *PlayerRepository) Update(ctx context.Context, companyID int64, player entities.Player, actorUserID int64) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin player update transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, strconv.FormatInt(actorUserID, 10)); err != nil {
		return fmt.Errorf("set audit actor for player update: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE players p
		SET
			name = $1,
			height = $2,
			weight = $3,
			position = $4,
			player_number = $5,
			profile_image_id = $6
		FROM teams t
		WHERE p.id = $7
		  AND p.team_id = $8
		  AND p.team_id = t.id
		  AND t.company_id = $9
		  AND t.deleted_at IS NULL
		  AND p.deleted_at IS NULL
	`, player.Name, player.Height, player.Weight, player.Position, player.PlayerNumber, player.ProfileImageID, player.ID, player.TeamID, companyID)
	if err != nil {
		return fmt.Errorf("update player: %w", classifyWriteError(err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read player update affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return playerdomain.ErrPlayerNotFound
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit player update transaction: %w", err)
	}

	return nil
}

func (r *PlayerRepository) SoftDelete(ctx context.Context, companyID, teamID, playerID, actorUserID int64) (deleted bool, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin player delete transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, strconv.FormatInt(actorUserID, 10)); err != nil {
		return false, fmt.Errorf("set audit actor for player delete: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE players p
		SET deleted_at = NOW()
		FROM teams t
		WHERE p.id = $1
		  AND p.team_id = $2
		  AND p.team_id = t.id
		  AND t.company_id = $3
		  AND t.deleted_at IS NULL
		  AND p.deleted_at IS NULL
	`, playerID, teamID, companyID)
	if err != nil {
		return false, fmt.Errorf("soft delete player: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read player delete affected rows: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("commit player delete transaction: %w", err)
	}

	return rowsAffected > 0, nil
}

type playerScanner interface {
	Scan(dest ...any) error
}

func scanPlayer(scanner playerScanner) (entities.Player, error) {
	var player entities.Player
	var profileImageID sql.NullInt64
	var deletedAt sql.NullTime

	err := scanner.Scan(
		&player.ID,
		&player.TeamID,
		&player.Name,
		&player.Height,
		&player.Weight,
		&player.Position,
		&player.PlayerNumber,
		&profileImageID,
		&player.CreatedAt,
		&player.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		return entities.Player{}, err
	}

	if profileImageID.Valid {
		id := profileImageID.Int64
		player.ProfileImageID = &id
	}
	if deletedAt.Valid {
		deletedTimestamp := deletedAt.Time
		player.DeletedAt = &deletedTimestamp
	}

	return player, nil
}

func (r *PlayerRepository) teamExists(ctx context.Context, companyID, teamID int64) bool {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM teams
			WHERE id = $1
			  AND company_id = $2
			  AND deleted_at IS NULL
		)
	`, teamID, companyID).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}

func classifyWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "uq_players_team_number_active":
				return playerdomain.ErrPlayerNumberAlreadyInUse
			case "uq_players_profile_image_id_active":
				return playerdomain.ErrPlayerProfileInUse
			}
		}
	}

	return err
}
