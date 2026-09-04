-- +goose Up
ALTER TABLE institutions ADD COLUMN IF NOT EXISTS logo_url VARCHAR(500);
CREATE INDEX IF NOT EXISTS idx_institutions_logo_url ON institutions(logo_url);

-- +goose Down
DROP INDEX IF EXISTS idx_institutions_logo_url;
ALTER TABLE institutions DROP COLUMN IF EXISTS logo_url;