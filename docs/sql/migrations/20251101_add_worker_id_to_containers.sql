-- Add worker_id to containers table to support distributed architecture
-- Nullable to maintain backward compatibility with existing local containers

ALTER TABLE containers
ADD COLUMN worker_id VARCHAR(64) REFERENCES workers(id) ON DELETE SET NULL;

CREATE INDEX idx_containers_worker_id ON containers(worker_id);

COMMENT ON COLUMN containers.worker_id IS 'Worker node that owns this container (NULL = local)';

