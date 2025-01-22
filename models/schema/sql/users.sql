CREATE TABLE "user" (
    id SMALLSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE,
    user_name VARCHAR(255) NOT NULL,
    unix_name VARCHAR(255) NOT NULL,
    user_type SMALLINT DEFAULT 1,
    forbidden BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_user_uuid_unix_name ON "user" (uuid, unix_name);


CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_user_updated_at
BEFORE UPDATE ON "user"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();