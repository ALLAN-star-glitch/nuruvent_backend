-- +goose Up
-- +goose StatementBegin

-- Rename title to name if needed
ALTER TABLE events RENAME COLUMN title TO name;

ALTER TABLE events ADD COLUMN IF NOT EXISTS display_name VARCHAR(150);

UPDATE events SET display_name = name WHERE display_name IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE events DROP COLUMN IF EXISTS display_name;
ALTER TABLE events RENAME COLUMN name TO title;

-- +goose StatementEnd