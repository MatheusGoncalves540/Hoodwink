-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_super_admin BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS request_logs (
    id BIGSERIAL PRIMARY KEY,
    at TIMESTAMPTZ NOT NULL DEFAULT now(),
    service TEXT NOT NULL CHECK (service IN ('backend', 'gameServer')),
    method TEXT,
    path TEXT,
    route TEXT,
    status INT,
    duration_ms INT,
    ip TEXT,
    user_agent TEXT,
    user_id UUID,
    username TEXT,
    request_id UUID,
    body TEXT,
    error TEXT
);

CREATE INDEX IF NOT EXISTS request_logs_at_idx ON request_logs (at DESC);
CREATE INDEX IF NOT EXISTS request_logs_route_idx ON request_logs (route);
CREATE INDEX IF NOT EXISTS request_logs_user_id_idx ON request_logs (user_id);

CREATE TABLE IF NOT EXISTS request_logs_archive (
    id BIGINT PRIMARY KEY,
    at TIMESTAMPTZ NOT NULL DEFAULT now(),
    service TEXT NOT NULL CHECK (service IN ('backend', 'gameServer')),
    method TEXT,
    path TEXT,
    route TEXT,
    status INT,
    duration_ms INT,
    ip TEXT,
    user_agent TEXT,
    user_id UUID,
    username TEXT,
    request_id UUID,
    body TEXT,
    error TEXT
);

CREATE INDEX IF NOT EXISTS request_logs_archive_at_idx ON request_logs_archive (at DESC);

-- +goose Down
DROP TABLE IF EXISTS request_logs_archive;
DROP TABLE IF EXISTS request_logs;
ALTER TABLE IF EXISTS users DROP COLUMN IF EXISTS is_super_admin;
