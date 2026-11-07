-- +goose Up
CREATE TABLE IF NOT EXISTS match_replays (
    match_id UUID PRIMARY KEY,
    room_id TEXT,
    engine_version INT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ NOT NULL,
    winner_id UUID,
    player_ids UUID[] NOT NULL,
    data BYTEA NOT NULL
);

CREATE INDEX IF NOT EXISTS match_replays_player_ids_gin ON match_replays USING GIN (player_ids);

-- +goose Down
DROP TABLE IF EXISTS match_replays;
