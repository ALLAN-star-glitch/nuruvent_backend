-- +goose Up
-- +goose StatementBegin

-- Add image_url column
ALTER TABLE events ADD COLUMN image_url VARCHAR(500);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop image_url column
ALTER TABLE events DROP COLUMN IF EXISTS image_url;

-- +goose StatementEnd