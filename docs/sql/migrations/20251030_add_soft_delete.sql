-- Add is_deleted column to containers table for soft delete
ALTER TABLE containers ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE NOT NULL;

-- Create index for better query performance
CREATE INDEX IF NOT EXISTS idx_containers_is_deleted ON containers(is_deleted);

-- Comment
COMMENT ON COLUMN containers.is_deleted IS 'Soft delete flag: true = deleted (hidden from user), false = active';

