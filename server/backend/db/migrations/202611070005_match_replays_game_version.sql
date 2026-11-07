-- +goose Up
-- Versão do jogo que gerou a partida (independente de engine_version, a versão
-- do motor de replay). Partidas antigas ficam com 'unknown'.
ALTER TABLE match_replays ADD COLUMN IF NOT EXISTS game_version TEXT NOT NULL DEFAULT 'unknown';

-- +goose Down
ALTER TABLE match_replays DROP COLUMN IF EXISTS game_version;
