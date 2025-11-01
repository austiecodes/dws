-- Migration: Add task_type field to tasks table
-- Purpose: Support CPU (concurrent) and GPU (exclusive) task scheduling
-- Date: 2025-10-30

-- 1. Add task_type column with default 'cpu' for backward compatibility
ALTER TABLE tasks 
ADD COLUMN task_type VARCHAR(10) NOT NULL DEFAULT 'cpu' 
CHECK (task_type IN ('cpu', 'gpu'));

-- 2. Add composite index for scheduler queries (type + status filtering)
CREATE INDEX idx_tasks_type_status ON tasks(task_type, status);

-- 3. Add index for priority-based dispatch within each type
CREATE INDEX idx_tasks_type_status_priority ON tasks(task_type, status, priority DESC, created_at ASC);

-- 4. Update column comment
COMMENT ON COLUMN tasks.task_type IS 'Task execution resource type: cpu (max 3 concurrent) or gpu (max 1 concurrent)';

-- Verification queries:
-- Check all existing tasks now have default 'cpu' type:
-- SELECT task_type, COUNT(*) FROM tasks GROUP BY task_type;

-- Check index creation:
-- SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'tasks' AND indexname LIKE '%type%';

