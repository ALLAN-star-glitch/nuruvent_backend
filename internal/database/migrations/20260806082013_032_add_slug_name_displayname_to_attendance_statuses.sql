-- +goose Up
-- +goose StatementBegin

-- Rename name to slug (since name currently holds the slug value)
ALTER TABLE attendance_statuses RENAME COLUMN name TO slug;

-- Add name column (canonical name)
ALTER TABLE attendance_statuses ADD COLUMN IF NOT EXISTS name VARCHAR(100);
UPDATE attendance_statuses SET name = initcap(slug) WHERE name IS NULL;

-- Add display_name column (UI display)
ALTER TABLE attendance_statuses ADD COLUMN IF NOT EXISTS display_name VARCHAR(150);
UPDATE attendance_statuses SET display_name = name WHERE display_name IS NULL;

-- Make name NOT NULL
ALTER TABLE attendance_statuses ALTER COLUMN name SET NOT NULL;

-- Add index on slug
CREATE INDEX IF NOT EXISTS idx_attendance_statuses_slug ON attendance_statuses (slug);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_attendance_statuses_slug;

ALTER TABLE attendance_statuses DROP COLUMN IF EXISTS display_name;
ALTER TABLE attendance_statuses DROP COLUMN IF EXISTS name;
ALTER TABLE attendance_statuses RENAME COLUMN slug TO name;

-- +goose StatementEnd