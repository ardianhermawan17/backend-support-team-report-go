package seeding

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"backend-sport-team-report-go/internal/platform/idgenerator"
	appcrypto "backend-sport-team-report-go/pkg/crypto"
)

const (
	userNodeID     int64 = 5
	companyNodeID  int64 = 6
	teamNodeID     int64 = 1
	playerNodeID   int64 = 2
	scheduleNodeID int64 = 3
	reportNodeID   int64 = 4
)

type Service struct {
	db            *sql.DB
	userIDs       idgenerator.IDGenerator
	companyIDs    idgenerator.IDGenerator
	teamIDs       idgenerator.IDGenerator
	playerIDs     idgenerator.IDGenerator
	scheduleIDs   idgenerator.IDGenerator
	reportIDs     idgenerator.IDGenerator
	adminPassword string
}

type seedTeam struct {
	Name                string
	FoundedYear         int
	HomebaseAddress     string
	HomebaseAddressCity string
}

type seedPlayer struct {
	TeamID       int64
	Name         string
	Height       float64
	Weight       float64
	Position     string
	PlayerNumber int
}

func NewService(db *sql.DB) (*Service, error) {
	userIDs, err := idgenerator.NewSnowflakeGenerator(userNodeID)
	if err != nil {
		return nil, fmt.Errorf("create user id generator: %w", err)
	}

	companyIDs, err := idgenerator.NewSnowflakeGenerator(companyNodeID)
	if err != nil {
		return nil, fmt.Errorf("create company id generator: %w", err)
	}

	teamIDs, err := idgenerator.NewSnowflakeGenerator(teamNodeID)
	if err != nil {
		return nil, fmt.Errorf("create team id generator: %w", err)
	}

	playerIDs, err := idgenerator.NewSnowflakeGenerator(playerNodeID)
	if err != nil {
		return nil, fmt.Errorf("create player id generator: %w", err)
	}

	scheduleIDs, err := idgenerator.NewSnowflakeGenerator(scheduleNodeID)
	if err != nil {
		return nil, fmt.Errorf("create schedule id generator: %w", err)
	}

	reportIDs, err := idgenerator.NewSnowflakeGenerator(reportNodeID)
	if err != nil {
		return nil, fmt.Errorf("create report id generator: %w", err)
	}

	return &Service{
		db:            db,
		userIDs:       userIDs,
		companyIDs:    companyIDs,
		teamIDs:       teamIDs,
		playerIDs:     playerIDs,
		scheduleIDs:   scheduleIDs,
		reportIDs:     reportIDs,
		adminPassword: "password",
	}, nil
}

func (s *Service) Seed(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin seeding transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	adminUserID, err := s.ensureUser(ctx, tx, "admin", "admin@gmail.com", s.adminPassword)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.user_id', $1, true)`, strconv.FormatInt(adminUserID, 10)); err != nil {
		return fmt.Errorf("set seeding audit actor: %w", err)
	}

	companyID, err := s.ensureCompany(ctx, tx, adminUserID, "Admin Soccer Company")
	if err != nil {
		return err
	}

	homeTeamID, err := s.ensureTeam(ctx, tx, companyID, seedTeam{
		Name:                "Admin United",
		FoundedYear:         2010,
		HomebaseAddress:     "123 Admin Street",
		HomebaseAddressCity: "Jakarta",
	})
	if err != nil {
		return err
	}

	guestTeamID, err := s.ensureTeam(ctx, tx, companyID, seedTeam{
		Name:                "Seeder City",
		FoundedYear:         2012,
		HomebaseAddress:     "456 Seeder Avenue",
		HomebaseAddressCity: "Bandung",
	})
	if err != nil {
		return err
	}

	homePlayerID, err := s.ensurePlayer(ctx, tx, seedPlayer{
		TeamID:       homeTeamID,
		Name:         "Admin Striker",
		Height:       182.5,
		Weight:       76.2,
		Position:     "striker",
		PlayerNumber: 9,
	})
	if err != nil {
		return err
	}

	if _, err = s.ensurePlayer(ctx, tx, seedPlayer{
		TeamID:       guestTeamID,
		Name:         "Seeder Goalkeeper",
		Height:       188.0,
		Weight:       80.4,
		Position:     "goalkeeper",
		PlayerNumber: 1,
	}); err != nil {
		return err
	}

	matchDate := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	matchTime := time.Date(1, time.January, 1, 19, 30, 0, 0, time.UTC)

	scheduleID, err := s.ensureSchedule(ctx, tx, companyID, homeTeamID, guestTeamID, matchDate, matchTime)
	if err != nil {
		return err
	}

	if _, err = s.ensureReport(ctx, tx, scheduleID, homeTeamID, guestTeamID, homePlayerID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit seeding transaction: %w", err)
	}

	return nil
}

func (s *Service) ensureUser(ctx context.Context, tx *sql.Tx, username, email, password string) (int64, error) {
	var (
		id           int64
		storedEmail  string
		passwordHash string
	)

	err := tx.QueryRowContext(ctx, `
		SELECT id, email, password_hash
		FROM users
		WHERE username = $1
		  AND deleted_at IS NULL
	`, username).Scan(&id, &storedEmail, &passwordHash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("lookup seeded user: %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		newID, genErr := s.userIDs.NewID()
		if genErr != nil {
			return 0, fmt.Errorf("generate user id: %w", genErr)
		}

		hash, hashErr := appcrypto.HashPassword(password)
		if hashErr != nil {
			return 0, fmt.Errorf("hash seeded admin password: %w", hashErr)
		}

		if _, execErr := tx.ExecContext(ctx, `
			INSERT INTO users (id, username, email, password_hash)
			VALUES ($1, $2, $3, $4)
		`, newID, username, email, hash); execErr != nil {
			return 0, fmt.Errorf("insert seeded user: %w", execErr)
		}

		return newID, nil
	}

	passwordMatches := appcrypto.VerifyPassword(password, passwordHash) == nil
	if storedEmail == email && passwordMatches {
		return id, nil
	}

	if !passwordMatches {
		var hashErr error
		passwordHash, hashErr = appcrypto.HashPassword(password)
		if hashErr != nil {
			return 0, fmt.Errorf("rehash seeded admin password: %w", hashErr)
		}
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE users
		SET email = $2,
		    password_hash = $3
		WHERE id = $1
	`, id, email, passwordHash); err != nil {
		return 0, fmt.Errorf("update seeded user: %w", err)
	}

	return id, nil
}

func (s *Service) ensureCompany(ctx context.Context, tx *sql.Tx, userID int64, name string) (int64, error) {
	var (
		id         int64
		storedName string
	)

	err := tx.QueryRowContext(ctx, `
		SELECT id, name
		FROM companies
		WHERE user_id = $1
		  AND deleted_at IS NULL
	`, userID).Scan(&id, &storedName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("lookup seeded company: %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		newID, genErr := s.companyIDs.NewID()
		if genErr != nil {
			return 0, fmt.Errorf("generate company id: %w", genErr)
		}

		if _, execErr := tx.ExecContext(ctx, `
			INSERT INTO companies (id, user_id, name)
			VALUES ($1, $2, $3)
		`, newID, userID, name); execErr != nil {
			return 0, fmt.Errorf("insert seeded company: %w", execErr)
		}

		return newID, nil
	}

	if storedName == name {
		return id, nil
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE companies
		SET name = $2
		WHERE id = $1
	`, id, name); err != nil {
		return 0, fmt.Errorf("update seeded company: %w", err)
	}

	return id, nil
}

func (s *Service) ensureTeam(ctx context.Context, tx *sql.Tx, companyID int64, team seedTeam) (int64, error) {
	var id int64

	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM teams
		WHERE company_id = $1
		  AND name = $2
		  AND deleted_at IS NULL
	`, companyID, team.Name).Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("lookup seeded team %s: %w", team.Name, err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		newID, genErr := s.teamIDs.NewID()
		if genErr != nil {
			return 0, fmt.Errorf("generate team id: %w", genErr)
		}

		if _, execErr := tx.ExecContext(ctx, `
			INSERT INTO teams (id, company_id, name, founded_year, homebase_address, city_of_homebase_address)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, newID, companyID, team.Name, team.FoundedYear, team.HomebaseAddress, team.HomebaseAddressCity); execErr != nil {
			return 0, fmt.Errorf("insert seeded team %s: %w", team.Name, execErr)
		}

		return newID, nil
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE teams
		SET founded_year = $2,
		    homebase_address = $3,
		    city_of_homebase_address = $4
		WHERE id = $1
	`, id, team.FoundedYear, team.HomebaseAddress, team.HomebaseAddressCity); err != nil {
		return 0, fmt.Errorf("update seeded team %s: %w", team.Name, err)
	}

	return id, nil
}

func (s *Service) ensurePlayer(ctx context.Context, tx *sql.Tx, player seedPlayer) (int64, error) {
	var id int64

	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM players
		WHERE team_id = $1
		  AND player_number = $2
		  AND deleted_at IS NULL
	`, player.TeamID, player.PlayerNumber).Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("lookup seeded player %s: %w", player.Name, err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		newID, genErr := s.playerIDs.NewID()
		if genErr != nil {
			return 0, fmt.Errorf("generate player id: %w", genErr)
		}

		if _, execErr := tx.ExecContext(ctx, `
			INSERT INTO players (id, team_id, name, height, weight, position, player_number, profile_image_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NULL)
		`, newID, player.TeamID, player.Name, player.Height, player.Weight, player.Position, player.PlayerNumber); execErr != nil {
			return 0, fmt.Errorf("insert seeded player %s: %w", player.Name, execErr)
		}

		return newID, nil
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE players
		SET name = $2,
		    height = $3,
		    weight = $4,
		    position = $5,
		    profile_image_id = NULL
		WHERE id = $1
	`, id, player.Name, player.Height, player.Weight, player.Position); err != nil {
		return 0, fmt.Errorf("update seeded player %s: %w", player.Name, err)
	}

	return id, nil
}

func (s *Service) ensureSchedule(ctx context.Context, tx *sql.Tx, companyID, homeTeamID, guestTeamID int64, matchDate, matchTime time.Time) (int64, error) {
	var id int64
	matchClock := matchTime.Format("15:04:05")

	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM schedules
		WHERE match_date = $1
		  AND match_time = $2
		  AND home_team_id = $3
		  AND guest_team_id = $4
		  AND deleted_at IS NULL
	`, matchDate, matchClock, homeTeamID, guestTeamID).Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("lookup seeded schedule: %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		newID, genErr := s.scheduleIDs.NewID()
		if genErr != nil {
			return 0, fmt.Errorf("generate schedule id: %w", genErr)
		}

		if _, execErr := tx.ExecContext(ctx, `
			INSERT INTO schedules (id, company_id, match_date, match_time, home_team_id, guest_team_id)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, newID, companyID, matchDate, matchClock, homeTeamID, guestTeamID); execErr != nil {
			return 0, fmt.Errorf("insert seeded schedule: %w", execErr)
		}

		return newID, nil
	}

	return id, nil
}

func (s *Service) ensureReport(ctx context.Context, tx *sql.Tx, scheduleID, homeTeamID, guestTeamID, topScorerID int64) (int64, error) {
	var id int64

	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM reports
		WHERE match_schedule_id = $1
		  AND deleted_at IS NULL
	`, scheduleID).Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("lookup seeded report: %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		newID, genErr := s.reportIDs.NewID()
		if genErr != nil {
			return 0, fmt.Errorf("generate report id: %w", genErr)
		}

		if _, execErr := tx.ExecContext(ctx, `
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
			) VALUES ($1, $2, $3, $4, 2, 1, 'home_team_win', $5, 1, 0)
		`, newID, scheduleID, homeTeamID, guestTeamID, topScorerID); execErr != nil {
			return 0, fmt.Errorf("insert seeded report: %w", execErr)
		}

		return newID, nil
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE reports
		SET home_team_id = $2,
		    guest_team_id = $3,
		    final_score_home = 2,
		    final_score_guest = 1,
		    status_match = 'home_team_win',
		    most_scoring_goal_player_id = $4,
		    accumulate_total_win_for_home_team_from_start_to_the_current_match_schedule = 1,
		    accumulate_total_win_for_guest_team_from_start_to_the_current_match_schedule = 0
		WHERE id = $1
	`, id, homeTeamID, guestTeamID, topScorerID); err != nil {
		return 0, fmt.Errorf("update seeded report: %w", err)
	}

	return id, nil
}
