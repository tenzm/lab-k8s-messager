-- +goose Up
ALTER TABLE messages ADD COLUMN IF NOT EXISTS file_name TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE messages DROP COLUMN IF EXISTS file_name;
