-- +goose Up
CREATE INDEX IF NOT EXISTS users_email_provider_idx ON users(email, provider);

-- +goose Down
DROP INDEX IF EXISTS users_email_provider_idx;
