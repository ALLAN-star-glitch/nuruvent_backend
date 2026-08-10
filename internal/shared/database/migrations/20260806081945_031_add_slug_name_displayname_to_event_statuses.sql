-- +goose Up
-- +goose StatementBegin

-- Rename name to slug (since name currently holds the slug value)
ALTER TABLE event_statuses RENAME COLUMN name TO slug;

-- Add name column (canonical name)
ALTER TABLE event_statuses ADD COLUMN IF NOT EXISTS name VARCHAR(100);
UPDATE event_statuses SET name = initcap(slug) WHERE name IS NULL;

-- Add display_name column (UI display)
ALTER TABLE event_statuses ADD COLUMN IF NOT EXISTS display_name VARCHAR(150);
UPDATE event_statuses SET display_name = name WHERE display_name IS NULL;

-- Make name NOT NULL
ALTER TABLE event_statuses ALTER COLUMN name SET NOT NULL;

-- Add index on slug
CREATE INDEX IF NOT EXISTS idx_event_statuses_slug ON event_statuses (slug);

-- Update existing data to match constants
-- Note: This assumes your existing data matches the constants

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_event_statuses_slug;

ALTER TABLE event_statuses DROP COLUMN IF EXISTS display_name;
ALTER TABLE event_statuses DROP COLUMN IF EXISTS name;
ALTER TABLE event_statuses RENAME COLUMN slug TO name;

-- +goose StatementEnd