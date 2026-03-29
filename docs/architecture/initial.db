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
        CREATE TYPE audit_action_type AS ENUM ('INSERT', 'UPDATE', 'SOFT_DELETE', 'RESTORE', 'DELETE');
    END IF;
END $$;

ALTER TYPE audit_action_type ADD VALUE IF NOT EXISTS 'SOFT_DELETE';
ALTER TYPE audit_action_type ADD VALUE IF NOT EXISTS 'RESTORE';

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS companies (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
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
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS teams (
    id BIGINT PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE ON UPDATE CASCADE,
    name VARCHAR(255) NOT NULL,
    logo_image_id BIGINT,
    founded_year INT,
    homebase_address TEXT,
    city_of_homebase_address VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
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
    profile_image_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
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
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_schedules_distinct_teams CHECK (home_team_id <> guest_team_id),
    CONSTRAINT ck_schedules_deleted_after_create CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE TABLE IF NOT EXISTS reports (
    id BIGINT PRIMARY KEY,
    match_schedule_id BIGINT NOT NULL REFERENCES schedules(id) ON DELETE CASCADE ON UPDATE CASCADE,
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
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_reports_score_nonnegative CHECK (final_score_home >= 0 AND final_score_guest >= 0),
    CONSTRAINT ck_reports_team_pair CHECK (home_team_id <> guest_team_id),
    CONSTRAINT ck_reports_deleted_after_create CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE images ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE teams ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE players ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE schedules ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE reports ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
ALTER TABLE companies DROP CONSTRAINT IF EXISTS companies_user_id_key;
ALTER TABLE companies DROP CONSTRAINT IF EXISTS companies_name_key;
ALTER TABLE images DROP CONSTRAINT IF EXISTS uq_images_polymorphic;
ALTER TABLE teams DROP CONSTRAINT IF EXISTS uq_teams_company_name;
ALTER TABLE teams DROP CONSTRAINT IF EXISTS teams_logo_image_id_key;
ALTER TABLE players DROP CONSTRAINT IF EXISTS uq_players_team_number;
ALTER TABLE players DROP CONSTRAINT IF EXISTS players_profile_image_id_key;
ALTER TABLE schedules DROP CONSTRAINT IF EXISTS uq_schedules_match;
ALTER TABLE reports DROP CONSTRAINT IF EXISTS reports_match_schedule_id_key;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_schedules_deleted_after_create') THEN
        ALTER TABLE schedules
        ADD CONSTRAINT ck_schedules_deleted_after_create CHECK (deleted_at IS NULL OR deleted_at >= created_at);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_reports_deleted_after_create') THEN
        ALTER TABLE reports
        ADD CONSTRAINT ck_reports_deleted_after_create CHECK (deleted_at IS NULL OR deleted_at >= created_at);
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_username_active ON users(username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_companies_user_id_active ON companies(user_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_companies_name_active ON companies(name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_images_polymorphic_active ON images(imageable_type, imageable_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_teams_logo_image_id_active ON teams(logo_image_id) WHERE deleted_at IS NULL AND logo_image_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_teams_company_name_active ON teams(company_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_players_profile_image_id_active ON players(profile_image_id) WHERE deleted_at IS NULL AND profile_image_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_players_team_number_active ON players(team_id, player_number) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_schedules_match_active ON schedules(match_date, match_time, home_team_id, guest_team_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_reports_match_schedule_active ON reports(match_schedule_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS logs (
    id BIGINT PRIMARY KEY DEFAULT nextval('logs_id_seq'),
    actor_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE,
    table_name VARCHAR(100) NOT NULL,
    record_id BIGINT NOT NULL,
    action audit_action_type NOT NULL,
    old_data JSONB,
    new_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE logs ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

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
    v_action audit_action_type;
BEGIN
    v_actor_user_id := NULLIF(current_setting('app.user_id', true), '')::BIGINT;

    IF TG_OP = 'INSERT' THEN
        INSERT INTO logs (actor_user_id, table_name, record_id, action, old_data, new_data, created_at)
        VALUES (v_actor_user_id, TG_TABLE_NAME, NEW.id, 'INSERT', NULL, to_jsonb(NEW), NOW());
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        v_action := 'UPDATE';

        IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
            v_action := 'SOFT_DELETE';
        ELSIF OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL THEN
            v_action := 'RESTORE';
        END IF;

        INSERT INTO logs (actor_user_id, table_name, record_id, action, old_data, new_data, created_at)
        VALUES (v_actor_user_id, TG_TABLE_NAME, NEW.id, v_action, to_jsonb(OLD), to_jsonb(NEW), NOW());
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO logs (actor_user_id, table_name, record_id, action, old_data, new_data, created_at)
        VALUES (v_actor_user_id, TG_TABLE_NAME, OLD.id, 'DELETE', to_jsonb(OLD), NULL, NOW());
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$;

CREATE OR REPLACE FUNCTION cascade_user_soft_delete()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        UPDATE companies
        SET deleted_at = NEW.deleted_at
        WHERE user_id = NEW.id
          AND deleted_at IS NULL;
    END IF;

    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION cascade_company_soft_delete()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        UPDATE teams
        SET deleted_at = NEW.deleted_at
        WHERE company_id = NEW.id
          AND deleted_at IS NULL;

        UPDATE schedules
        SET deleted_at = NEW.deleted_at
        WHERE company_id = NEW.id
          AND deleted_at IS NULL;
    END IF;

    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION cascade_team_soft_delete()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        UPDATE players
        SET deleted_at = NEW.deleted_at
        WHERE team_id = NEW.id
          AND deleted_at IS NULL;

        UPDATE schedules
        SET deleted_at = NEW.deleted_at
        WHERE (home_team_id = NEW.id OR guest_team_id = NEW.id)
          AND deleted_at IS NULL;

        UPDATE images
        SET deleted_at = NEW.deleted_at
        WHERE imageable_type = 'team'
          AND imageable_id = NEW.id
          AND deleted_at IS NULL;
    END IF;

    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION cascade_player_soft_delete()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        UPDATE images
        SET deleted_at = NEW.deleted_at
        WHERE imageable_type = 'player'
          AND imageable_id = NEW.id
          AND deleted_at IS NULL;
    END IF;

    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION cascade_schedule_soft_delete()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        UPDATE reports
        SET deleted_at = NEW.deleted_at
        WHERE match_schedule_id = NEW.id
          AND deleted_at IS NULL;
    END IF;

    RETURN NEW;
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
    IF NEW.deleted_at IS NOT NULL THEN
        RETURN NEW;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM companies
        WHERE id = NEW.company_id
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'Schedule company must be active';
    END IF;

    SELECT company_id INTO v_home_company_id FROM teams WHERE id = NEW.home_team_id AND deleted_at IS NULL;
    SELECT company_id INTO v_guest_company_id FROM teams WHERE id = NEW.guest_team_id AND deleted_at IS NULL;

    IF v_home_company_id IS NULL OR v_guest_company_id IS NULL THEN
        RAISE EXCEPTION 'Both teams must exist and be active';
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
    IF NEW.deleted_at IS NOT NULL THEN
        RETURN NEW;
    END IF;

    SELECT home_team_id, guest_team_id
    INTO v_home_team_id, v_guest_team_id
    FROM schedules
    WHERE id = NEW.match_schedule_id
      AND deleted_at IS NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Match schedule does not exist or is deleted';
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

DROP TRIGGER IF EXISTS trg_users_soft_delete_cascade ON users;
CREATE TRIGGER trg_users_soft_delete_cascade
AFTER UPDATE OF deleted_at ON users
FOR EACH ROW EXECUTE FUNCTION cascade_user_soft_delete();

DROP TRIGGER IF EXISTS trg_companies_soft_delete_cascade ON companies;
CREATE TRIGGER trg_companies_soft_delete_cascade
AFTER UPDATE OF deleted_at ON companies
FOR EACH ROW EXECUTE FUNCTION cascade_company_soft_delete();

DROP TRIGGER IF EXISTS trg_teams_soft_delete_cascade ON teams;
CREATE TRIGGER trg_teams_soft_delete_cascade
AFTER UPDATE OF deleted_at ON teams
FOR EACH ROW EXECUTE FUNCTION cascade_team_soft_delete();

DROP TRIGGER IF EXISTS trg_players_soft_delete_cascade ON players;
CREATE TRIGGER trg_players_soft_delete_cascade
AFTER UPDATE OF deleted_at ON players
FOR EACH ROW EXECUTE FUNCTION cascade_player_soft_delete();

DROP TRIGGER IF EXISTS trg_schedules_soft_delete_cascade ON schedules;
CREATE TRIGGER trg_schedules_soft_delete_cascade
AFTER UPDATE OF deleted_at ON schedules
FOR EACH ROW EXECUTE FUNCTION cascade_schedule_soft_delete();

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
