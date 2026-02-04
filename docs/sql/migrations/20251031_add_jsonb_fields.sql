-- Migration: Add JSONB fields for flexible metadata storage
-- Purpose: Leverage PostgreSQL JSONB for structured configuration and extensibility
-- Date: 2025-10-31

-- ============================================================================
-- 1. Containers: Store creation configuration (env vars, passwords, etc.)
-- ============================================================================

ALTER TABLE containers 
ADD COLUMN config JSONB NOT NULL DEFAULT '{}';

-- GIN index for efficient JSONB queries (e.g., key existence, containment)
CREATE INDEX idx_containers_config_gin ON containers USING GIN (config);

COMMENT ON COLUMN containers.config IS 'Container creation configuration in JSON format. Stores env vars, SSH password, restart policy, labels, etc. Enables container recreation with identical settings.';

-- Example queries:
-- Find containers with specific env var: 
--   SELECT * FROM containers WHERE config->'env' @> '["GPU_ENABLED=true"]';
-- Find containers with labels:
--   SELECT * FROM containers WHERE config->'labels' ? 'team';

-- ============================================================================
-- 2. Tasks: Store extensible metadata (resources, retry config, labels)
-- ============================================================================

ALTER TABLE tasks 
ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}';

-- GIN index for metadata queries
CREATE INDEX idx_tasks_metadata_gin ON tasks USING GIN (metadata);

COMMENT ON COLUMN tasks.metadata IS 'Extensible task metadata in JSON format. Can store resource requirements, retry configuration, dependencies, labels, and custom attributes without schema changes.';

-- Example queries:
-- Find tasks by team label:
--   SELECT * FROM tasks WHERE metadata->'labels'->>'team' = 'ml';
-- Find GPU tasks:
--   SELECT * FROM tasks WHERE metadata->'resources' ? 'gpu';
-- Find tasks with dependencies:
--   SELECT * FROM tasks WHERE metadata ? 'dependencies';

-- ============================================================================
-- 3. Users: Store user preferences and settings
-- ============================================================================

ALTER TABLE users 
ADD COLUMN preferences JSONB NOT NULL DEFAULT '{}';

-- GIN index for preference lookups
CREATE INDEX idx_users_preferences_gin ON users USING GIN (preferences);

COMMENT ON COLUMN users.preferences IS 'User preferences and settings in JSON format. Stores notification preferences, UI settings (theme, language), API tokens, and other user-specific configurations.';

-- Example queries:
-- Find users with email notifications enabled:
--   SELECT * FROM users WHERE preferences->'notifications'->>'email' = 'true';
-- Find users with dark theme:
--   SELECT * FROM users WHERE preferences->'ui'->>'theme' = 'dark';

-- ============================================================================
-- Verification queries
-- ============================================================================

-- Check columns added successfully:
-- SELECT column_name, data_type FROM information_schema.columns 
-- WHERE table_name IN ('containers', 'tasks', 'users') 
--   AND column_name IN ('config', 'metadata', 'preferences');

-- Check GIN indexes created:
-- SELECT indexname, indexdef FROM pg_indexes 
-- WHERE indexname LIKE '%_gin';

