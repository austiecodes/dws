-- Migration: Add RBAC support
-- This migration adds the role column to users table and migrates existing is_admin data

-- Add role column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';

-- Create index on role column for faster queries
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Migrate existing data:
-- 1. First admin user becomes super_admin
-- 2. Other admins become admin
-- 3. Everyone else keeps 'user' (default)

-- Step 1: Find the first admin and make them super_admin
UPDATE users 
SET role = 'super_admin' 
WHERE id = (
    SELECT id FROM users 
    WHERE is_admin = true 
    ORDER BY created_at ASC 
    LIMIT 1
);

-- Step 2: Make remaining admins into 'admin' role
UPDATE users 
SET role = 'admin' 
WHERE is_admin = true 
  AND role = 'user';

-- Note: The casbin_rule table will be auto-created by gorm-adapter
-- when the RBAC enforcer is initialized. No manual table creation needed.

-- Add comment to mark is_admin as deprecated
COMMENT ON COLUMN users.is_admin IS 'Deprecated: Use role column instead. Kept for rollback safety.';
