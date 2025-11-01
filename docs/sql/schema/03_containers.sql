-- Containers table - Docker container lifecycle management
-- This is the authoritative, up-to-date schema definition
-- Last updated: 2025-11-01

CREATE TABLE IF NOT EXISTS containers (
    id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE,
    
    -- Foreign keys
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    worker_id VARCHAR(64) REFERENCES workers(id) ON DELETE SET NULL,
    
    -- Docker identity
    container_id TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    image TEXT NOT NULL,
    
    -- Network
    host_ssh_port INTEGER NOT NULL UNIQUE,
    
    -- State
    status TEXT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    
    -- Configuration
    config JSONB NOT NULL DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_containers_user_id ON containers(user_id);
CREATE INDEX IF NOT EXISTS idx_containers_worker_id ON containers(worker_id);
CREATE INDEX IF NOT EXISTS idx_containers_is_deleted ON containers(is_deleted);
CREATE INDEX IF NOT EXISTS idx_containers_worker_status ON containers(worker_id, status) WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_containers_config_gin ON containers USING GIN (config);

-- Trigger function for automatic timestamp updates
CREATE OR REPLACE FUNCTION set_containers_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update updated_at on row modification
CREATE TRIGGER set_containers_timestamp
    BEFORE UPDATE ON containers
    FOR EACH ROW
    EXECUTE FUNCTION set_containers_timestamp();

-- Comments
COMMENT ON TABLE containers IS 'User-owned Docker containers for isolated development environments';
COMMENT ON COLUMN containers.uuid IS 'Public-facing UUID for API references';
COMMENT ON COLUMN containers.user_id IS 'Owner of this container';
COMMENT ON COLUMN containers.worker_id IS 'Worker node that owns this container (NULL = local/unscheduled)';
COMMENT ON COLUMN containers.container_id IS 'Docker container ID (long format)';
COMMENT ON COLUMN containers.name IS 'User-defined container name';
COMMENT ON COLUMN containers.image IS 'Docker image used to create the container';
COMMENT ON COLUMN containers.host_ssh_port IS 'Host port mapped to container SSH (22000-22999 range)';
COMMENT ON COLUMN containers.status IS 'Container state: running, stopped, created, exited, etc.';
COMMENT ON COLUMN containers.is_deleted IS 'Soft delete flag: true = deleted (hidden from user), false = active';
COMMENT ON COLUMN containers.config IS 'Container creation configuration in JSON format (env vars, SSH password, restart policy, labels, resource limits)';
