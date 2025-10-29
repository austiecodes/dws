-- Tasks table for job scheduling system
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    container_id INTEGER NOT NULL REFERENCES containers(id) ON DELETE RESTRICT,
    command TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'killed')),
    expected_duration INTEGER NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    output TEXT,
    exit_code INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_container_id ON tasks(container_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_status_priority ON tasks(status, priority DESC);
CREATE INDEX idx_tasks_started_at ON tasks(started_at) WHERE started_at IS NOT NULL;

-- Comments
COMMENT ON TABLE tasks IS 'User-submitted jobs to be executed in containers';
COMMENT ON COLUMN tasks.expected_duration IS 'Expected duration in seconds (for user notification)';
COMMENT ON COLUMN tasks.priority IS 'Higher priority tasks are scheduled first (default 0)';
COMMENT ON COLUMN tasks.output IS 'Combined stdout/stderr from task execution';

