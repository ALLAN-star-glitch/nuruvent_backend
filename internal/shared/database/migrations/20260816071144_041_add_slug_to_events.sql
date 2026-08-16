-- +goose Up
-- +goose StatementBegin

-- 1. Add slug column (NOT NULL directly since no existing data)
ALTER TABLE events ADD COLUMN slug VARCHAR(255) NOT NULL;

-- 2. Add unique constraint
ALTER TABLE events ADD CONSTRAINT events_slug_unique UNIQUE (slug);

-- 3. Add index for slug lookups
CREATE INDEX idx_events_slug ON events(slug);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 1. Drop index
DROP INDEX IF EXISTS idx_events_slug;

-- 2. Drop unique constraint
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_slug_unique;

-- 3. Drop slug column
ALTER TABLE events DROP COLUMN IF EXISTS slug;

-- +goose StatementEnd