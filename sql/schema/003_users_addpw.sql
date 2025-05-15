-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS pw_hash TEXT NOT NULL DEFAULT 'unset';

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS pw_hash;