-- internal/shared/database/migrations/XXX_remove_type_column_from_events.up.sql

-- +goose Up
-- +goose StatementBegin

-- Remove the redundant type column
ALTER TABLE events DROP COLUMN IF EXISTS type;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Add back the type column if rolling back
ALTER TABLE events ADD COLUMN type VARCHAR(50);

-- +goose StatementEnd