-- +goose Up
-- Chave fixa para casar com BACKEND_API_KEY do gameServer/.env.example (dev).
INSERT INTO api_keys (apikey, system_name)
VALUES ('00000000-0000-0000-0000-000000000000', 'gameServer');

-- +goose Down
DELETE FROM api_keys WHERE system_name = 'gameServer';