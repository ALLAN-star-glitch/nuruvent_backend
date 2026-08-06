-- +goose Up
-- +goose StatementBegin

-- Check if slug column exists, if not add it
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'account_types' AND column_name = 'slug'
    ) THEN
        ALTER TABLE account_types ADD COLUMN slug VARCHAR(50) UNIQUE;
        UPDATE account_types SET slug = LOWER(REPLACE(name, ' ', '-'));
        ALTER TABLE account_types ALTER COLUMN slug SET NOT NULL;
        CREATE INDEX idx_account_types_slug ON account_types (slug);
    END IF;
END $$;

-- Check if display_name column exists, if not add it
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'account_types' AND column_name = 'display_name'
    ) THEN
        ALTER TABLE account_types ADD COLUMN display_name VARCHAR(150);
        UPDATE account_types SET display_name = name WHERE display_name IS NULL;
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_account_types_slug;

ALTER TABLE account_types DROP COLUMN IF EXISTS display_name;
ALTER TABLE account_types DROP COLUMN IF EXISTS slug;

-- +goose StatementEnd