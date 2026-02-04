-- Add lease-based liveness columns to workers table
ALTER TABLE workers
  ADD COLUMN IF NOT EXISTS lease_expires_at timestamptz NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS heartbeat_ttl_secs int NOT NULL DEFAULT 30;

-- Optional: backfill lease_expires_at to last_heartbeat if present
UPDATE workers
SET lease_expires_at = COALESCE(last_heartbeat, now())
WHERE lease_expires_at IS DISTINCT FROM COALESCE(last_heartbeat, now());

-- Index to speed up queries that compute online workers by lease
CREATE INDEX IF NOT EXISTS idx_workers_lease_expires
ON workers (lease_expires_at);

-- View for convenient API reads (optional in your environment)
-- DROP VIEW IF EXISTS v_workers_status;
CREATE OR REPLACE VIEW v_workers_status AS
SELECT
  w.*,
  (now() < w.lease_expires_at AND w.status <> 'maintenance') AS is_online
FROM workers w;

