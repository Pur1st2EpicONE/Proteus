-- +goose Up
CREATE TABLE IF NOT EXISTS images (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    uuid        VARCHAR(36) NOT NULL UNIQUE,
    object_key VARCHAR(255) NOT NULL UNIQUE,
    processed_key VARCHAR(255),
    status      VARCHAR(30) NOT NULL DEFAULT 'processing',
    created_at  TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS images;