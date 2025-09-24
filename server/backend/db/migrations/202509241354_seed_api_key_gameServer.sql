-- +goose Up
INSERT INTO api_keys (apikey, system_name)
VALUES (gen_random_uuid(), 'gameServer');

-- +goose Down
DELETE FROM api_keys WHERE system_name = 'gameServer';