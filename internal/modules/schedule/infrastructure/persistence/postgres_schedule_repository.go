package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	scheduledomain "backend-sport-team-report-go/internal/modules/schedule/domain"
	"backend-sport-team-report-go/internal/modules/schedule/domain/entities"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/paginator"
)

type ScheduleRepository struct {
	db *sql.DB
}

func NewScheduleRepository(conn *postgres.Connection) *ScheduleRepository {
	return &ScheduleRepository{db: conn.DB()}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule entities.Schedule, actorUserID int64) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin schedule create transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, strconv.FormatInt(actorUserID, 10)); err != nil {
		return fmt.Errorf("set audit actor for schedule create: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO schedules (id, company_id, match_date, match_time, home_team_id, guest_team_id)
		SELECT $1, $2, $3, $4, home_team.id, guest_team.id
		FROM teams home_team
		JOIN teams guest_team ON guest_team.id = $6
		WHERE home_team.id = $5
		  AND home_team.company_id = $2
		  AND guest_team.company_id = $2
		  AND home_team.deleted_at IS NULL
		  AND guest_team.deleted_at IS NULL
	`, schedule.ID, schedule.CompanyID, schedule.MatchDate, schedule.MatchTime, schedule.HomeTeamID, schedule.GuestTeamID)
	if err != nil {
		return fmt.Errorf("insert schedule: %w", classifyWriteError(err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read schedule create affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return scheduledomain.ErrScheduleTeamNotFound
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit schedule create transaction: %w", err)
	}

	return nil
}

func (r *ScheduleRepository) ListByCompany(ctx context.Context, companyID int64, params paginator.Params) (paginator.Result[entities.Schedule], error) {
	var totalItems int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM schedules
		WHERE company_id = $1
		  AND deleted_at IS NULL
	`, companyID).Scan(&totalItems); err != nil {
		return paginator.Result[entities.Schedule]{}, fmt.Errorf("count schedules by company: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			company_id,
			match_date,
			match_time,
			home_team_id,
			guest_team_id,
			created_at,
			updated_at,
			deleted_at
		FROM schedules
		WHERE company_id = $1
		  AND deleted_at IS NULL
		ORDER BY match_date ASC, match_time ASC, created_at ASC, id ASC
		LIMIT $2
		OFFSET $3
	`, companyID, params.Limit, params.Offset)
	if err != nil {
		return paginator.Result[entities.Schedule]{}, fmt.Errorf("list schedules by company: %w", err)
	}
	defer rows.Close()

	schedules := make([]entities.Schedule, 0)
	for rows.Next() {
		schedule, scanErr := scanSchedule(rows)
		if scanErr != nil {
			return paginator.Result[entities.Schedule]{}, scanErr
		}
		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		return paginator.Result[entities.Schedule]{}, fmt.Errorf("iterate schedules by company: %w", err)
	}

	return paginator.Result[entities.Schedule]{
		Items: schedules,
		Meta:  paginator.BuildMeta(params, totalItems),
	}, nil
}

func (r *ScheduleRepository) FindByIDAndCompany(ctx context.Context, scheduleID, companyID int64) (entities.Schedule, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			company_id,
			match_date,
			match_time,
			home_team_id,
			guest_team_id,
			created_at,
			updated_at,
			deleted_at
		FROM schedules
		WHERE id = $1
		  AND company_id = $2
		  AND deleted_at IS NULL
	`, scheduleID, companyID)

	schedule, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.Schedule{}, scheduledomain.ErrScheduleNotFound
		}
		return entities.Schedule{}, fmt.Errorf("find schedule by id and company: %w", err)
	}

	return schedule, nil
}

func (r *ScheduleRepository) Update(ctx context.Context, schedule entities.Schedule, actorUserID int64) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin schedule update transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, strconv.FormatInt(actorUserID, 10)); err != nil {
		return fmt.Errorf("set audit actor for schedule update: %w", err)
	}

	var currentUpdatedAt time.Time
	if err = tx.QueryRowContext(ctx, `
		SELECT updated_at
		FROM schedules
		WHERE id = $1
		  AND company_id = $2
		  AND deleted_at IS NULL
		FOR UPDATE
	`, schedule.ID, schedule.CompanyID).Scan(&currentUpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return scheduledomain.ErrScheduleNotFound
		}
		return fmt.Errorf("lock schedule for update: %w", err)
	}

	if !currentUpdatedAt.Equal(schedule.UpdatedAt) {
		return scheduledomain.ErrScheduleConcurrentModification
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE schedules s
		SET
			match_date = $1,
			match_time = $2,
			home_team_id = home_team.id,
			guest_team_id = guest_team.id
		FROM teams home_team
		JOIN teams guest_team ON guest_team.id = $4
		WHERE s.id = $5
		  AND s.company_id = $6
		  AND s.deleted_at IS NULL
		  AND home_team.id = $3
		  AND home_team.company_id = $6
		  AND guest_team.company_id = $6
		  AND home_team.deleted_at IS NULL
		  AND guest_team.deleted_at IS NULL
	`, schedule.MatchDate, schedule.MatchTime, schedule.HomeTeamID, schedule.GuestTeamID, schedule.ID, schedule.CompanyID)
	if err != nil {
		return fmt.Errorf("update schedule: %w", classifyWriteError(err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read schedule update affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return scheduledomain.ErrScheduleTeamNotFound
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit schedule update transaction: %w", err)
	}

	return nil
}

func (r *ScheduleRepository) SoftDelete(ctx context.Context, scheduleID, companyID, actorUserID int64) (deleted bool, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin schedule delete transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, strconv.FormatInt(actorUserID, 10)); err != nil {
		return false, fmt.Errorf("set audit actor for schedule delete: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE schedules
		SET deleted_at = NOW()
		WHERE id = $1
		  AND company_id = $2
		  AND deleted_at IS NULL
	`, scheduleID, companyID)
	if err != nil {
		return false, fmt.Errorf("soft delete schedule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read schedule delete affected rows: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("commit schedule delete transaction: %w", err)
	}

	return rowsAffected > 0, nil
}

type scheduleScanner interface {
	Scan(dest ...any) error
}

func scanSchedule(scanner scheduleScanner) (entities.Schedule, error) {
	var schedule entities.Schedule
	var deletedAt sql.NullTime
	var matchTimeRaw string

	err := scanner.Scan(
		&schedule.ID,
		&schedule.CompanyID,
		&schedule.MatchDate,
		&matchTimeRaw,
		&schedule.HomeTeamID,
		&schedule.GuestTeamID,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		return entities.Schedule{}, err
	}

	matchTime, err := time.Parse("15:04:05", matchTimeRaw)
	if err != nil {
		return entities.Schedule{}, fmt.Errorf("parse match_time: %w", err)
	}
	schedule.MatchTime = matchTime

	if deletedAt.Valid {
		deletedTimestamp := deletedAt.Time
		schedule.DeletedAt = &deletedTimestamp
	}

	return schedule, nil
}

func (r *ScheduleRepository) scheduleExists(ctx context.Context, scheduleID, companyID int64) bool {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM schedules
			WHERE id = $1
			  AND company_id = $2
			  AND deleted_at IS NULL
		)
	`, scheduleID, companyID).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}

func classifyWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if pgErr.ConstraintName == "uq_schedules_match_active" {
				return scheduledomain.ErrScheduleAlreadyExists
			}
		case "40001", "40P01":
			return scheduledomain.ErrScheduleConcurrentModification
		}
	}

	return err
}
