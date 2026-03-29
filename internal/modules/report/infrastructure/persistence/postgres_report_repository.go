package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	reportdomain "backend-sport-team-report-go/internal/modules/report/domain"
	"backend-sport-team-report-go/internal/modules/report/domain/entities"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/paginator"
)

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(conn *postgres.Connection) *ReportRepository {
	return &ReportRepository{db: conn.DB()}
}

func (r *ReportRepository) Create(ctx context.Context, report entities.Report, actorUserID int64) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin report create transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, strconv.FormatInt(actorUserID, 10)); err != nil {
		return fmt.Errorf("set audit actor for report create: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		WITH schedule_snapshot AS (
			SELECT s.id, s.home_team_id, s.guest_team_id
			FROM schedules s
			JOIN teams home_team ON home_team.id = s.home_team_id
			JOIN teams guest_team ON guest_team.id = s.guest_team_id
			WHERE s.id = $2
			  AND s.company_id = $3
			  AND s.deleted_at IS NULL
			  AND home_team.deleted_at IS NULL
			  AND guest_team.deleted_at IS NULL
		), valid_top_scorer AS (
			SELECT 1 AS ok
			WHERE $7::BIGINT IS NULL
			UNION ALL
			SELECT 1
			FROM players p
			JOIN schedule_snapshot ss ON TRUE
			WHERE p.id = $7
			  AND p.deleted_at IS NULL
			  AND p.team_id IN (ss.home_team_id, ss.guest_team_id)
			LIMIT 1
		)
		INSERT INTO reports (
			id,
			match_schedule_id,
			home_team_id,
			guest_team_id,
			final_score_home,
			final_score_guest,
			status_match,
			most_scoring_goal_player_id,
			accumulate_total_win_for_home_team_from_start_to_the_current_match_schedule,
			accumulate_total_win_for_guest_team_from_start_to_the_current_match_schedule
		)
		SELECT
			$1,
			ss.id,
			ss.home_team_id,
			ss.guest_team_id,
			$4,
			$5,
			$6,
			$7,
			0,
			0
		FROM schedule_snapshot ss
		WHERE EXISTS (SELECT 1 FROM valid_top_scorer)
	`,
		report.ID,
		report.MatchScheduleID,
		report.CompanyID,
		report.FinalScoreHome,
		report.FinalScoreGuest,
		report.StatusMatch,
		report.MostScoringGoalPlayerID,
	)
	if err != nil {
		return fmt.Errorf("insert report: %w", classifyWriteError(err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read report create affected rows: %w", err)
	}
	if rowsAffected == 0 {
		if !r.scheduleExists(ctx, tx, report.MatchScheduleID, report.CompanyID) {
			return reportdomain.ErrReportScheduleNotFound
		}
		return reportdomain.ErrReportTopScorerNotFound
	}

	if err = r.recomputeAccumulateWins(ctx, tx, report.CompanyID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit report create transaction: %w", err)
	}

	return nil
}

func (r *ReportRepository) ListByCompany(ctx context.Context, companyID int64, params paginator.Params) (paginator.Result[entities.Report], error) {
	var totalItems int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM reports r
		JOIN schedules s ON s.id = r.match_schedule_id
		WHERE s.company_id = $1
		  AND r.deleted_at IS NULL
		  AND s.deleted_at IS NULL
	`, companyID).Scan(&totalItems); err != nil {
		return paginator.Result[entities.Report]{}, fmt.Errorf("count reports by company: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			r.id,
			s.company_id,
			r.match_schedule_id,
			r.home_team_id,
			r.guest_team_id,
			r.final_score_home,
			r.final_score_guest,
			r.status_match,
			r.most_scoring_goal_player_id,
			r.accumulate_total_win_for_home_team_from_start_to_the_current_match_schedule,
			r.accumulate_total_win_for_guest_team_from_start_to_the_current_match_schedule,
			r.created_at,
			r.updated_at,
			r.deleted_at,
			s.match_date,
			s.match_time
		FROM reports r
		JOIN schedules s ON s.id = r.match_schedule_id
		WHERE s.company_id = $1
		  AND r.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		ORDER BY s.match_date ASC, s.match_time ASC, r.created_at ASC, r.id ASC
		LIMIT $2
		OFFSET $3
	`, companyID, params.Limit, params.Offset)
	if err != nil {
		return paginator.Result[entities.Report]{}, fmt.Errorf("list reports by company: %w", err)
	}
	defer rows.Close()

	reports := make([]entities.Report, 0)
	for rows.Next() {
		report, scanErr := scanReport(rows)
		if scanErr != nil {
			return paginator.Result[entities.Report]{}, scanErr
		}
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return paginator.Result[entities.Report]{}, fmt.Errorf("iterate reports by company: %w", err)
	}

	return paginator.Result[entities.Report]{
		Items: reports,
		Meta:  paginator.BuildMeta(params, totalItems),
	}, nil
}

func (r *ReportRepository) FindByIDAndCompany(ctx context.Context, reportID, companyID int64) (entities.Report, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			r.id,
			s.company_id,
			r.match_schedule_id,
			r.home_team_id,
			r.guest_team_id,
			r.final_score_home,
			r.final_score_guest,
			r.status_match,
			r.most_scoring_goal_player_id,
			r.accumulate_total_win_for_home_team_from_start_to_the_current_match_schedule,
			r.accumulate_total_win_for_guest_team_from_start_to_the_current_match_schedule,
			r.created_at,
			r.updated_at,
			r.deleted_at,
			s.match_date,
			s.match_time
		FROM reports r
		JOIN schedules s ON s.id = r.match_schedule_id
		WHERE r.id = $1
		  AND s.company_id = $2
		  AND r.deleted_at IS NULL
		  AND s.deleted_at IS NULL
	`, reportID, companyID)

	report, err := scanReport(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.Report{}, reportdomain.ErrReportNotFound
		}
		return entities.Report{}, fmt.Errorf("find report by id and company: %w", err)
	}

	return report, nil
}

func (r *ReportRepository) Update(ctx context.Context, report entities.Report, actorUserID int64) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin report update transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, strconv.FormatInt(actorUserID, 10)); err != nil {
		return fmt.Errorf("set audit actor for report update: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		WITH schedule_snapshot AS (
			SELECT s.id, s.home_team_id, s.guest_team_id
			FROM schedules s
			JOIN teams home_team ON home_team.id = s.home_team_id
			JOIN teams guest_team ON guest_team.id = s.guest_team_id
			WHERE s.id = $3
			  AND s.company_id = $2
			  AND s.deleted_at IS NULL
			  AND home_team.deleted_at IS NULL
			  AND guest_team.deleted_at IS NULL
		), valid_top_scorer AS (
			SELECT 1 AS ok
			WHERE $7::BIGINT IS NULL
			UNION ALL
			SELECT 1
			FROM players p
			JOIN schedule_snapshot ss ON TRUE
			WHERE p.id = $7
			  AND p.deleted_at IS NULL
			  AND p.team_id IN (ss.home_team_id, ss.guest_team_id)
			LIMIT 1
		)
		UPDATE reports r
		SET
			match_schedule_id = ss.id,
			home_team_id = ss.home_team_id,
			guest_team_id = ss.guest_team_id,
			final_score_home = $4,
			final_score_guest = $5,
			status_match = $6,
			most_scoring_goal_player_id = $7
		FROM schedule_snapshot ss
		WHERE r.id = $1
		  AND r.deleted_at IS NULL
		  AND EXISTS (SELECT 1 FROM valid_top_scorer)
	`,
		report.ID,
		report.CompanyID,
		report.MatchScheduleID,
		report.FinalScoreHome,
		report.FinalScoreGuest,
		report.StatusMatch,
		report.MostScoringGoalPlayerID,
	)
	if err != nil {
		return fmt.Errorf("update report: %w", classifyWriteError(err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read report update affected rows: %w", err)
	}
	if rowsAffected == 0 {
		if !r.reportExists(ctx, tx, report.ID, report.CompanyID) {
			return reportdomain.ErrReportNotFound
		}
		if !r.scheduleExists(ctx, tx, report.MatchScheduleID, report.CompanyID) {
			return reportdomain.ErrReportScheduleNotFound
		}
		return reportdomain.ErrReportTopScorerNotFound
	}

	if err = r.recomputeAccumulateWins(ctx, tx, report.CompanyID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit report update transaction: %w", err)
	}

	return nil
}

func (r *ReportRepository) SoftDelete(ctx context.Context, reportID, companyID, actorUserID int64) (deleted bool, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin report delete transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, strconv.FormatInt(actorUserID, 10)); err != nil {
		return false, fmt.Errorf("set audit actor for report delete: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE reports r
		SET deleted_at = NOW()
		FROM schedules s
		WHERE r.id = $1
		  AND r.match_schedule_id = s.id
		  AND s.company_id = $2
		  AND r.deleted_at IS NULL
		  AND s.deleted_at IS NULL
	`, reportID, companyID)
	if err != nil {
		return false, fmt.Errorf("soft delete report: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read report delete affected rows: %w", err)
	}

	if rowsAffected > 0 {
		if err = r.recomputeAccumulateWins(ctx, tx, companyID); err != nil {
			return false, err
		}
	}

	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("commit report delete transaction: %w", err)
	}

	return rowsAffected > 0, nil
}

func (r *ReportRepository) recomputeAccumulateWins(ctx context.Context, tx *sql.Tx, companyID int64) error {
	_, err := tx.ExecContext(ctx, `
		WITH company_reports AS (
			SELECT
				r.id,
				r.home_team_id,
				r.guest_team_id,
				r.status_match,
				s.match_date,
				s.match_time
			FROM reports r
			JOIN schedules s ON s.id = r.match_schedule_id
			WHERE s.company_id = $1
			  AND r.deleted_at IS NULL
			  AND s.deleted_at IS NULL
		)
		UPDATE reports target
		SET
			accumulate_total_win_for_home_team_from_start_to_the_current_match_schedule = COALESCE((
				SELECT COUNT(*)
				FROM company_reports previous
				WHERE (
					previous.match_date < current_row.match_date
					OR (previous.match_date = current_row.match_date AND previous.match_time < current_row.match_time)
					OR (previous.match_date = current_row.match_date AND previous.match_time = current_row.match_time AND previous.id <= current_row.id)
				)
				AND (
					(previous.status_match = 'home_team_win' AND previous.home_team_id = current_row.home_team_id)
					OR (previous.status_match = 'guest_team_win' AND previous.guest_team_id = current_row.home_team_id)
				)
			), 0),
			accumulate_total_win_for_guest_team_from_start_to_the_current_match_schedule = COALESCE((
				SELECT COUNT(*)
				FROM company_reports previous
				WHERE (
					previous.match_date < current_row.match_date
					OR (previous.match_date = current_row.match_date AND previous.match_time < current_row.match_time)
					OR (previous.match_date = current_row.match_date AND previous.match_time = current_row.match_time AND previous.id <= current_row.id)
				)
				AND (
					(previous.status_match = 'home_team_win' AND previous.home_team_id = current_row.guest_team_id)
					OR (previous.status_match = 'guest_team_win' AND previous.guest_team_id = current_row.guest_team_id)
				)
			), 0)
		FROM company_reports current_row
		WHERE target.id = current_row.id
	`, companyID)
	if err != nil {
		return fmt.Errorf("recompute report accumulated wins: %w", err)
	}

	return nil
}

type reportScanner interface {
	Scan(dest ...any) error
}

func scanReport(scanner reportScanner) (entities.Report, error) {
	var report entities.Report
	var deletedAt sql.NullTime
	var mostScoringGoalPlayerID sql.NullInt64
	var matchTimeRaw string
	var statusMatch string
	var matchDate time.Time

	err := scanner.Scan(
		&report.ID,
		&report.CompanyID,
		&report.MatchScheduleID,
		&report.HomeTeamID,
		&report.GuestTeamID,
		&report.FinalScoreHome,
		&report.FinalScoreGuest,
		&statusMatch,
		&mostScoringGoalPlayerID,
		&report.AccumulateTotalWinForHomeTeamFromStartToTheCurrentMatchSchedule,
		&report.AccumulateTotalWinForGuestTeamFromStartToTheCurrentMatchSchedule,
		&report.CreatedAt,
		&report.UpdatedAt,
		&deletedAt,
		&matchDate,
		&matchTimeRaw,
	)
	if err != nil {
		return entities.Report{}, err
	}

	report.StatusMatch = statusMatch

	if mostScoringGoalPlayerID.Valid {
		value := mostScoringGoalPlayerID.Int64
		report.MostScoringGoalPlayerID = &value
	}

	if deletedAt.Valid {
		deletedTimestamp := deletedAt.Time
		report.DeletedAt = &deletedTimestamp
	}

	if _, err := time.Parse("15:04:05", matchTimeRaw); err != nil {
		return entities.Report{}, fmt.Errorf("parse match_time: %w", err)
	}

	_ = matchDate

	return report, nil
}

func (r *ReportRepository) reportExists(ctx context.Context, tx *sql.Tx, reportID, companyID int64) bool {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM reports r
			JOIN schedules s ON s.id = r.match_schedule_id
			WHERE r.id = $1
			  AND s.company_id = $2
			  AND r.deleted_at IS NULL
			  AND s.deleted_at IS NULL
		)
	`, reportID, companyID).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}

func (r *ReportRepository) scheduleExists(ctx context.Context, tx *sql.Tx, scheduleID, companyID int64) bool {
	var exists bool
	err := tx.QueryRowContext(ctx, `
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
		if pgErr.Code == "23505" && pgErr.ConstraintName == "uq_reports_match_schedule_active" {
			return reportdomain.ErrReportAlreadyExists
		}
	}

	return err
}
