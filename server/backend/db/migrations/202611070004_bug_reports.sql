-- +goose Up
CREATE TABLE IF NOT EXISTS bug_reports (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    username TEXT,
    source TEXT NOT NULL CHECK (source IN ('partida', 'lobby')),
    match_id UUID,
    turns INT[] NOT NULL DEFAULT '{}',
    category TEXT CHECK (category IN ('moedas', 'carta', 'travou', 'regra', 'tela', 'outro')),
    description TEXT CHECK (description IS NULL OR char_length(description) <= 2000),
    context JSONB CHECK (context IS NULL OR octet_length(context::text) <= 8192),
    ip TEXT,
    user_agent TEXT,
    status TEXT NOT NULL DEFAULT 'novo' CHECK (status IN ('novo', 'visto', 'resolvido')),
    CONSTRAINT bug_reports_turns_max CHECK (array_length(turns, 1) <= 10)
);

CREATE INDEX IF NOT EXISTS bug_reports_created_at_idx ON bug_reports (created_at DESC);
CREATE INDEX IF NOT EXISTS bug_reports_status_idx ON bug_reports (status);
CREATE INDEX IF NOT EXISTS bug_reports_match_id_idx ON bug_reports (match_id);
CREATE UNIQUE INDEX IF NOT EXISTS bug_reports_user_match_uniq ON bug_reports (user_id, match_id)
    WHERE user_id IS NOT NULL AND match_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS bug_reports;
