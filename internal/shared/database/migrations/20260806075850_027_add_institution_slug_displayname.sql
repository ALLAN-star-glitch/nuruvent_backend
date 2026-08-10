-- +goose Up
-- +goose StatementBegin

ALTER TABLE institutions ADD COLUMN IF NOT EXISTS slug VARCHAR(50) UNIQUE;
ALTER TABLE institutions ADD COLUMN IF NOT EXISTS display_name VARCHAR(150);

UPDATE institutions SET slug = LOWER(REPLACE(name, ' ', '-'));
UPDATE institutions SET display_name = name WHERE display_name IS NULL;

ALTER TABLE institutions ALTER COLUMN slug SET NOT NULL;

CREATE INDEX idx_institutions_slug ON institutions (slug);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_institutions_slug;
ALTER TABLE institutions DROP COLUMN IF EXISTS slug;
ALTER TABLE institutions DROP COLUMN IF EXISTS display_name;

-- +goose StatementEnd