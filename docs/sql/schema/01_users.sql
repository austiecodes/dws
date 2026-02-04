-- Users table - Core user accounts for the platform
-- This is the authoritative, up-to-date schema definition
-- Last updated: 2025-10-30

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    display_name VARCHAR(120) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT false,
    preferences JSONB NOT NULL DEFAULT '{}',
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- GIN index for JSONB preference queries
CREATE INDEX IF NOT EXISTS idx_users_preferences_gin ON users USING GIN (preferences);

-- Trigger function for automatic timestamp updates
CREATE OR REPLACE FUNCTION set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update updated_at on row modification
CREATE TRIGGER set_users_timestamp
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_timestamp();

-- Comments
COMMENT ON TABLE users IS 'Platform user accounts with authentication and role management';
COMMENT ON COLUMN users.email IS 'Unique email address for login';
COMMENT ON COLUMN users.display_name IS 'User-friendly name displayed in UI';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password';
COMMENT ON COLUMN users.is_admin IS 'Admin flag for privilege escalation';
COMMENT ON COLUMN users.preferences IS 'User preferences and settings in JSON format (notifications, UI theme, API tokens, etc.)';
COMMENT ON COLUMN users.last_login_at IS 'Timestamp of most recent successful login';

