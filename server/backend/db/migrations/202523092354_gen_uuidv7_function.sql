-- -- +goose Up
-- CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- CREATE OR REPLACE FUNCTION uuidv7()
-- RETURNS uuid
-- LANGUAGE SQL
-- AS $$
--     SELECT (
--         lpad(to_hex((EXTRACT(EPOCH FROM clock_timestamp()) * 1000)::BIGINT), 12, '0') ||
--         lpad(to_hex((get_byte(r,0)::int << 8 | get_byte(r,1)::int) & 0x0fff | 0x7000), 4, '0') ||
--         lpad(to_hex((get_byte(r,2)::int << 8 | get_byte(r,3)::int) & 0x3fff | 0x8000), 4, '0') ||
--         encode(substring(r from 5), 'hex')
--     )::uuid
--     FROM gen_random_bytes(10) r;
-- $$;

-- -- +goose Down
-- DROP FUNCTION IF EXISTS uuidv7();
