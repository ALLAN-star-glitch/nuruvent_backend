-- +goose Up
-- +goose StatementBegin

-- Add slug column to businesses
ALTER TABLE businesses ADD COLUMN slug VARCHAR(255) NOT NULL DEFAULT '';

-- Add unique constraint
ALTER TABLE businesses ADD CONSTRAINT businesses_slug_unique UNIQUE (slug);

-- Create index for faster lookups
CREATE INDEX idx_businesses_slug ON businesses(slug);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop index
DROP INDEX IF EXISTS idx_businesses_slug;

-- Drop unique constraint
ALTER TABLE businesses DROP CONSTRAINT IF EXISTS businesses_slug_unique;

-- Drop slug column
ALTER TABLE businesses DROP COLUMN IF EXISTS slug;

-- +goose StatementEnd