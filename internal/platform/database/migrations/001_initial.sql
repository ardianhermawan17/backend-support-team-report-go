-- soccer-team-report / PostgreSQL initial schema
-- Application should generate Snowflake IDs for business tables.
-- logs.id uses a BIGINT sequence so audit rows are always writable by triggers.

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'position_type') THEN
        CREATE TYPE position_type AS ENUM ('striker', 'midfielder', 'defender', 'goalkeeper');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'match_status_type') THEN
        CREATE TYPE match_status_type AS ENUM ('home_team_win', 'guest_team_win', 'draw');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'audit_action_type') THEN
        CREATE TYPE audit_action_type AS ENUM ('INSERT', 'UPDATE', 'DELETE');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS companies (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE SEQUENCE IF NOT EXISTS logs_id_seq AS BIGINT;

CREATE TABLE IF NOT EXISTS images (
    id BIGINT PRIMARY KEY,
    imageable_id BIGINT NOT NULL,
    imageable_type VARCHAR(50) NOT NULL,
    url TEXT,
    mime_type VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_images_polymorphic UNIQUE (imageable_type, imageable_id)
);

CREATE TABLE IF NOT EXISTS teams (
    id BIGINT PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE ON UPDATE CASCADE,
    name VARCHAR(255) NOT NULL,
    logo_image_id BIGINT UNIQUE,
    founded_year INT,
    homebase_address TEXT,
    city_of_homebase_address VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_teams_company_name UNIQUE (company_id, name),
    CONSTRAINT fk_teams_logo_image FOREIGN KEY (logo_image_id) REFERENCES images(id) ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS players (
    id BIGINT PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE ON UPDATE CASCADE,
    name VARCHAR(255) NOT NULL,
    height NUMERIC(5,2),
    weight NUMERIC(5,2),
    position position_type NOT NULL,
    player_number INT NOT NULL,
    profile_image_id BIGINT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_players_team_number UNIQUE (team_id, player_number),
    CONSTRAINT fk_players_profile_image FOREIGN KEY (profile_image_id) REFERENCES images(id) ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS schedules (
    id BIGINT PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE ON UPDATE CASCADE,
    match_date DATE NOT NULL,
    match_time TIME NOT NULL,
    home_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    guest_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_schedules_distinct_teams CHECK (home_team_id <> guest_team_id),
    CONSTRAINT uq_schedules_match UNIQUE (match_date, match_time, home_team_id, guest_team_id)
);

CREATE TABLE IF NOT EXISTS reports (
    id BIGINT PRIMARY KEY,
    match_schedule_id BIGINT NOT NULL UNIQUE REFERENCES schedules(id) ON DELETE CASCADE ON UPDATE CASCADE,
    home_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    guest_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    final_score_home INT NOT NULL DEFAULT 0,
    final_score_guest INT NOT NULL DEFAULT 0,
    status_match match_status_type NOT NULL,
    most_scoring_goal_player_id BIGINT REFERENCES players(id) ON DELETE SET NULL ON UPDATE CASCADE,
    accumulate_total_win_for_home_team_from_start_to_the_current_match_schedule INT NOT NULL DEFAULT 0,
    accumulate_total_win_for_guest_team_from_start_to_the_current_match_schedule INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_reports_score_nonnegative CHECK (final_score_home >= 0 AND final_score_guest >= 0),
    CONSTRAINT ck_reports_team_pair CHECK (home_team_id <> guest_team_id)
);

CREATE TABLE IF NOT EXISTS logs (
    id BIGINT PRIMARY KEY DEFAULT nextval('logs_id_seq'),
    actor_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE,
    table_name VARCHAR(100) NOT NULL,
    record_id BIGINT NOT NULL,
    action audit_action_type NOT NULL,
    old_data JSONB,
    new_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_companies_user_id ON companies(user_id);
CREATE INDEX IF NOT EXISTS idx_teams_company_id ON teams(company_id);
CREATE INDEX IF NOT EXISTS idx_players_team_id ON players(team_id);
CREATE INDEX IF NOT EXISTS idx_schedules_company_id ON schedules(company_id);
CREATE INDEX IF NOT EXISTS idx_schedules_home_team_id ON schedules(home_team_id);
CREATE INDEX IF NOT EXISTS idx_schedules_guest_team_id ON schedules(guest_team_id);
CREATE INDEX IF NOT EXISTS idx_reports_match_schedule_id ON reports(match_schedule_id);
CREATE INDEX IF NOT EXISTS idx_logs_actor_user_id ON logs(actor_user_id);
CREATE INDEX IF NOT EXISTS idx_logs_table_record ON logs(table_name, record_id);
CREATE INDEX IF NOT EXISTS idx_images_polymorphic ON images(imageable_type, imageable_id);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION audit_row_changes()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_actor_user_id BIGINT;
BEGIN
    v_actor_user_id := NULLIF(current_setting('app.user_id', true), '')::BIGINT;

    IF TG_OP = 'INSERT' THEN
        INSERT INTO logs (actor_user_id, table_name, record_id, action, old_data, new_data, created_at)
        VALUES (v_actor_user_id, TG_TABLE_NAME, NEW.id, 'INSERT', NULL, to_jsonb(NEW), NOW());
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO logs (actor_user_id, table_name, record_id, action, old_data, new_data, created_at)
        VALUES (v_actor_user_id, TG_TABLE_NAME, NEW.id, 'UPDATE', to_jsonb(OLD), to_jsonb(NEW), NOW());
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO logs (actor_user_id, table_name, record_id, action, old_data, new_data, created_at)
        VALUES (v_actor_user_id, TG_TABLE_NAME, OLD.id, 'DELETE', to_jsonb(OLD), NULL, NOW());
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$;

-- Validation triggers for same-company match scheduling and consistent report snapshot.
CREATE OR REPLACE FUNCTION validate_schedule_company()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_home_company_id BIGINT;
    v_guest_company_id BIGINT;
BEGIN
    SELECT company_id INTO v_home_company_id FROM teams WHERE id = NEW.home_team_id;
    SELECT company_id INTO v_guest_company_id FROM teams WHERE id = NEW.guest_team_id;

    IF v_home_company_id IS NULL OR v_guest_company_id IS NULL THEN
        RAISE EXCEPTION 'Both teams must exist';
    END IF;

    IF v_home_company_id <> NEW.company_id OR v_guest_company_id <> NEW.company_id THEN
        RAISE EXCEPTION 'Schedule teams must belong to the same company as the schedule';
    END IF;

    IF NEW.home_team_id = NEW.guest_team_id THEN
        RAISE EXCEPTION 'Home team and guest team must be different';
    END IF;

    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION validate_report_snapshot()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_home_team_id BIGINT;
    v_guest_team_id BIGINT;
BEGIN
    SELECT home_team_id, guest_team_id
    INTO v_home_team_id, v_guest_team_id
    FROM schedules
    WHERE id = NEW.match_schedule_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Match schedule does not exist';
    END IF;

    IF NEW.home_team_id <> v_home_team_id OR NEW.guest_team_id <> v_guest_team_id THEN
        RAISE EXCEPTION 'Report teams must match the schedule teams';
    END IF;

    IF NEW.home_team_id = NEW.guest_team_id THEN
        RAISE EXCEPTION 'Home team and guest team must be different';
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;
CREATE TRIGGER trg_users_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_companies_set_updated_at ON companies;
CREATE TRIGGER trg_companies_set_updated_at
BEFORE UPDATE ON companies
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_images_set_updated_at ON images;
CREATE TRIGGER trg_images_set_updated_at
BEFORE UPDATE ON images
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_teams_set_updated_at ON teams;
CREATE TRIGGER trg_teams_set_updated_at
BEFORE UPDATE ON teams
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_players_set_updated_at ON players;
CREATE TRIGGER trg_players_set_updated_at
BEFORE UPDATE ON players
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_schedules_set_updated_at ON schedules;
CREATE TRIGGER trg_schedules_set_updated_at
BEFORE UPDATE ON schedules
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_reports_set_updated_at ON reports;
CREATE TRIGGER trg_reports_set_updated_at
BEFORE UPDATE ON reports
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_validate_schedule_company ON schedules;
CREATE TRIGGER trg_validate_schedule_company
BEFORE INSERT OR UPDATE ON schedules
FOR EACH ROW EXECUTE FUNCTION validate_schedule_company();

DROP TRIGGER IF EXISTS trg_validate_report_snapshot ON reports;
CREATE TRIGGER trg_validate_report_snapshot
BEFORE INSERT OR UPDATE ON reports
FOR EACH ROW EXECUTE FUNCTION validate_report_snapshot();

DROP TRIGGER IF EXISTS trg_audit_users ON users;
CREATE TRIGGER trg_audit_users
AFTER INSERT OR UPDATE OR DELETE ON users
FOR EACH ROW EXECUTE FUNCTION audit_row_changes();

DROP TRIGGER IF EXISTS trg_audit_companies ON companies;
CREATE TRIGGER trg_audit_companies
AFTER INSERT OR UPDATE OR DELETE ON companies
FOR EACH ROW EXECUTE FUNCTION audit_row_changes();

DROP TRIGGER IF EXISTS trg_audit_images ON images;
CREATE TRIGGER trg_audit_images
AFTER INSERT OR UPDATE OR DELETE ON images
FOR EACH ROW EXECUTE FUNCTION audit_row_changes();

DROP TRIGGER IF EXISTS trg_audit_teams ON teams;
CREATE TRIGGER trg_audit_teams
AFTER INSERT OR UPDATE OR DELETE ON teams
FOR EACH ROW EXECUTE FUNCTION audit_row_changes();

DROP TRIGGER IF EXISTS trg_audit_players ON players;
CREATE TRIGGER trg_audit_players
AFTER INSERT OR UPDATE OR DELETE ON players
FOR EACH ROW EXECUTE FUNCTION audit_row_changes();

DROP TRIGGER IF EXISTS trg_audit_schedules ON schedules;
CREATE TRIGGER trg_audit_schedules
AFTER INSERT OR UPDATE OR DELETE ON schedules
FOR EACH ROW EXECUTE FUNCTION audit_row_changes();

DROP TRIGGER IF EXISTS trg_audit_reports ON reports;
CREATE TRIGGER trg_audit_reports
AFTER INSERT OR UPDATE OR DELETE ON reports
FOR EACH ROW EXECUTE FUNCTION audit_row_changes();

COMMIT;
