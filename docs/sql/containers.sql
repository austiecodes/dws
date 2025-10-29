CREATE TABLE IF NOT EXISTS containers (
    id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE,
    container_id TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    image TEXT NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    host_ssh_port INTEGER NOT NULL UNIQUE,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION set_containers_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'set_containers_timestamp') THEN
        CREATE TRIGGER set_containers_timestamp
        BEFORE UPDATE ON containers
        FOR EACH ROW
        EXECUTE FUNCTION set_containers_timestamp();
    END IF;
END $$;
