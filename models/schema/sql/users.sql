CREATE TABLE "user" (
    id SERIAL PRIMARY KEY,
    uuid INT NOT NULL UNIQUE,
    unix_name VARCHAR(255) NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    forbidden BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建 uuid 和 unix_name 的联合索引
CREATE INDEX idx_user_uuid_unix_name ON "user" (uuid, unix_name);

-- 触发器
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