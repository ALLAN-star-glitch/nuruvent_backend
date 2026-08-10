-- +goose Up
-- +goose StatementBegin

ALTER TABLE accounts ADD COLUMN IF NOT EXISTS slug VARCHAR(50) UNIQUE;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS display_name VARCHAR(150);

UPDATE accounts SET slug = LOWER(REPLACE(name, ' ', '-'));
UPDATE accounts SET display_name = name WHERE display_name IS NULL;

ALTER TABLE accounts ALTER COLUMN slug SET NOT NULL;

CREATE INDEX idx_accounts_slug ON accounts (slug);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_accounts_slug;
ALTER TABLE accounts DROP COLUMN IF EXISTS slug;
ALTER TABLE accounts DROP COLUMN IF EXISTS display_name;

-- +goose StatementEnd