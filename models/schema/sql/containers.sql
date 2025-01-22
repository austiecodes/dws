CREATE TABLE "container" (
    id SMALLSERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    container_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    image VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_container_uuid_container_id ON "container" (uuid, container_id);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_container_updated_at
BEFORE UPDATE ON "container"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();