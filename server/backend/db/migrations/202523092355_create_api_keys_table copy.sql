-- +goose Up
CREATE TABLE api_keys (
    apikey UUID PRIMARY KEY,
    system_name VARCHAR(255) UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS api_keys;