-- Workers table: manages distributed worker nodes
-- Each worker represents a physical machine running Docker containers
-- Last updated: 2025-11-01

CREATE TABLE IF NOT EXISTS workers (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'offline' 
        CHECK (status IN ('online', 'offline', 'maintenance')),
    
    -- Capacity limits (NULL = unlimited)
    max_containers INTEGER,
    max_cpu_cores INTEGER,
    max_memory_gb INTEGER,
    max_gpu_count INTEGER,
    
    -- Runtime state
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    
    -- Extensible metadata
    metadata JSONB NOT NULL DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_workers_status ON workers(status);
CREATE INDEX IF NOT EXISTS idx_workers_last_heartbeat ON workers(last_heartbeat);
CREATE INDEX IF NOT EXISTS idx_workers_metadata_gin ON workers USING GIN (metadata);

-- Trigger for automatic timestamp updates
CREATE OR REPLACE FUNCTION set_workers_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_workers_timestamp
    BEFORE UPDATE ON workers
    FOR EACH ROW
    EXECUTE FUNCTION set_workers_timestamp();

-- Comments
COMMENT ON TABLE workers IS 'Distributed worker nodes that run Docker containers and execute tasks';
COMMENT ON COLUMN workers.id IS 'Unique worker identifier (e.g., worker-gpu-01), set by operator';
COMMENT ON COLUMN workers.name IS 'Human-readable name for the worker';
COMMENT ON COLUMN workers.address IS 'gRPC endpoint address (host:port)';
COMMENT ON COLUMN workers.status IS 'Current worker status: online, offline, maintenance';
COMMENT ON COLUMN workers.max_containers IS 'Maximum number of containers this worker can run (NULL = unlimited)';
COMMENT ON COLUMN workers.max_cpu_cores IS 'Maximum CPU cores available (NULL = unlimited)';
COMMENT ON COLUMN workers.max_memory_gb IS 'Maximum memory in GB available (NULL = unlimited)';
COMMENT ON COLUMN workers.max_gpu_count IS 'Number of GPUs available (NULL = no GPU)';
COMMENT ON COLUMN workers.last_heartbeat IS 'Timestamp of last heartbeat from worker';
COMMENT ON COLUMN workers.metadata IS 'Extensible JSON metadata: region, zone, tags, custom attributes';
