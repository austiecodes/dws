CREATE TABLE image (
    id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    image_id VARCHAR(255) NOT NULL,
    repository VARCHAR(255) NOT NULL,
    tag VARCHAR(255) NOT NULL,
    created VARCHAR(255) NOT NULL,
    size VARCHAR(255) NOT NULL
);

CREATE INDEX idx_image_uuid_image_id ON image (uuid, image_id);