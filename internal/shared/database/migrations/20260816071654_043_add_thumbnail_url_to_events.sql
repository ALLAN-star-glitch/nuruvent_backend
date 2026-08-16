-- +goose Up
-- +goose StatementBegin

-- Add thumbnail_url column
ALTER TABLE events ADD COLUMN thumbnail_url VARCHAR(500);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop thumbnail_url column
ALTER TABLE events DROP COLUMN IF EXISTS thumbnail_url;

-- +goose StatementEnd