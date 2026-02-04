-- Tasks table - Job scheduling and execution tracking
-- This is the authoritative, up-to-date schema definition
-- Last updated: 2025-11-01

CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    
    -- Foreign keys
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    container_id INTEGER NOT NULL REFERENCES containers(id) ON DELETE RESTRICT,
    worker_id VARCHAR(64) REFERENCES workers(id) ON DELETE SET NULL,
    
    -- Task definition
    command TEXT NOT NULL,
    task_type VARCHAR(10) NOT NULL DEFAULT 'cpu' 
        CHECK (task_type IN ('cpu', 'gpu')),
    expected_duration INTEGER NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    
    -- State
    status VARCHAR(20) NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending', 'running', 'completed', 'failed', 'killed')),
    
    -- Execution tracking
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Results
    output TEXT,
    exit_code INTEGER,
    
    -- Extensible metadata
    metadata JSONB NOT NULL DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_container_id ON tasks(container_id);
CREATE INDEX IF NOT EXISTS idx_tasks_worker_id ON tasks(worker_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_status_priority ON tasks(status, priority DESC);
CREATE INDEX IF NOT EXISTS idx_tasks_started_at ON tasks(started_at) WHERE started_at IS NOT NULL;

-- Indexes for task type scheduling
CREATE INDEX IF NOT EXISTS idx_tasks_type_status ON tasks(task_type, status);
CREATE INDEX IF NOT EXISTS idx_tasks_type_status_priority ON tasks(task_type, status, priority DESC, created_at ASC);

-- Indexes for worker-specific queries
CREATE INDEX IF NOT EXISTS idx_tasks_worker_status ON tasks(worker_id, status) WHERE worker_id IS NOT NULL;

-- GIN index for JSONB metadata queries
CREATE INDEX IF NOT EXISTS idx_tasks_metadata_gin ON tasks USING GIN (metadata);

-- Trigger for automatic timestamp updates
CREATE OR REPLACE FUNCTION set_tasks_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_tasks_timestamp
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION set_tasks_timestamp();

-- Comments
COMMENT ON TABLE tasks IS 'User-submitted jobs to be executed in containers on worker nodes';
COMMENT ON COLUMN tasks.user_id IS 'Owner of this task';
COMMENT ON COLUMN tasks.container_id IS 'Container where this task will execute';
COMMENT ON COLUMN tasks.worker_id IS 'Worker node executing this task (derived from container, cached for performance)';
COMMENT ON COLUMN tasks.command IS 'Shell command to execute in the container';
COMMENT ON COLUMN tasks.task_type IS 'Task execution resource type: cpu (max 3 concurrent) or gpu (max 1 concurrent)';
COMMENT ON COLUMN tasks.expected_duration IS 'Expected duration in seconds (for user notification)';
COMMENT ON COLUMN tasks.priority IS 'Higher priority tasks are scheduled first (default 0)';
COMMENT ON COLUMN tasks.status IS 'Task lifecycle state: pending → running → completed/failed/killed';
COMMENT ON COLUMN tasks.started_at IS 'Timestamp when task execution began';
COMMENT ON COLUMN tasks.completed_at IS 'Timestamp when task finished (success or failure)';
COMMENT ON COLUMN tasks.output IS 'Combined stdout/stderr from task execution';
COMMENT ON COLUMN tasks.exit_code IS 'Process exit code (0 = success, non-zero = error)';
COMMENT ON COLUMN tasks.metadata IS 'Extensible task metadata in JSON format (resource limits, retry config, dependencies, labels)';
