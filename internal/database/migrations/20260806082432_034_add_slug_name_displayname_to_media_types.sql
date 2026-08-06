-- +goose Up
-- +goose StatementBegin

-- Check if slug column exists, if not add it
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'media_types' AND column_name = 'slug'
    ) THEN
        ALTER TABLE media_types ADD COLUMN slug VARCHAR(50) UNIQUE;
        UPDATE media_types SET slug = LOWER(REPLACE(name, ' ', '-'));
        ALTER TABLE media_types ALTER COLUMN slug SET NOT NULL;
        CREATE INDEX idx_media_types_slug ON media_types (slug);
    END IF;
END $$;

-- Check if name column exists, if not add it
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'media_types' AND column_name = 'name'
    ) THEN
        ALTER TABLE media_types ADD COLUMN name VARCHAR(100);
        UPDATE media_types SET name = display_name WHERE name IS NULL;
        ALTER TABLE media_types ALTER COLUMN name SET NOT NULL;
    END IF;
END $$;

-- Check if display_name column exists, if not add it
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'media_types' AND column_name = 'display_name'
    ) THEN
        ALTER TABLE media_types ADD COLUMN display_name VARCHAR(150);
        UPDATE media_types SET display_name = name WHERE display_name IS NULL;
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_media_types_slug;

ALTER TABLE media_types DROP COLUMN IF EXISTS display_name;
ALTER TABLE media_types DROP COLUMN IF EXISTS name;
ALTER TABLE media_types DROP COLUMN IF EXISTS slug;

-- +goose StatementEnd