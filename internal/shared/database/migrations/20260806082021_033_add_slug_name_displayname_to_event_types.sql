-- +goose Up
-- +goose StatementBegin

-- Check if name column exists, if not add it
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'event_types' AND column_name = 'name'
    ) THEN
        ALTER TABLE event_types ADD COLUMN name VARCHAR(100);
        UPDATE event_types SET name = display_name WHERE name IS NULL;
        ALTER TABLE event_types ALTER COLUMN name SET NOT NULL;
    END IF;
END $$;

-- Check if display_name column exists, if not add it
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'event_types' AND column_name = 'display_name'
    ) THEN
        ALTER TABLE event_types ADD COLUMN display_name VARCHAR(150);
        UPDATE event_types SET display_name = name WHERE display_name IS NULL;
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE event_types DROP COLUMN IF EXISTS display_name;
ALTER TABLE event_types DROP COLUMN IF EXISTS name;

-- +goose StatementEnd