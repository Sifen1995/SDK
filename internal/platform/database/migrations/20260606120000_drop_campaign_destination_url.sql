-- +goose Up
ALTER TABLE campaigns DROP COLUMN IF EXISTS destination_url;

-- +goose Down
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS destination_url TEXT;
