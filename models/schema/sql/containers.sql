CREATE TABLE "container" (
    id SERIAL PRIMARY KEY,
    uuid INT NOT NULL,
    container_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    image VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建 uuid 和 container_id 的联合索引
CREATE INDEX idx_container_uuid_container_id ON "container" (uuid, container_id);

-- 触发器
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