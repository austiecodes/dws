CREATE EXTENSION IF NOT EXISTS citext;

-- Users table for the DWS platform.
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email CITEXT NOT NULL,
    display_name VARCHAR(120) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_email_key UNIQUE (email)
);

-- Optional trigger to keep updated_at fresh.
CREATE OR REPLACE FUNCTION set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'set_users_timestamp') THEN
        CREATE TRIGGER set_users_timestamp
        BEFORE UPDATE ON users
        FOR EACH ROW
        EXECUTE FUNCTION set_timestamp();
    END IF;
END $$;

-- Seed admin example (replace PASSWORD_HASH before running).
-- INSERT INTO users (email, display_name, password_hash, is_admin)
-- VALUES ('admin@example.com', 'Lab Admin', '<bcrypt hash>', TRUE);
