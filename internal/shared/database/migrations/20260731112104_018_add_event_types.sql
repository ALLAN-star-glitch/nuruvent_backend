-- +goose Up
-- +goose StatementBegin

-- ================================================
-- STEP 1: Create event_types table
-- ================================================
CREATE TABLE IF NOT EXISTS event_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(20),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    supports_certificate BOOLEAN DEFAULT true,
    min_duration INTEGER DEFAULT 60,
    max_duration INTEGER DEFAULT 480,
    meta_title VARCHAR(150),
    meta_description TEXT,
    slug VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- ================================================
-- STEP 2: Add indexes
-- ================================================
CREATE INDEX idx_event_types_name ON event_types (name);
CREATE INDEX idx_event_types_slug ON event_types (slug);
CREATE INDEX idx_event_types_is_active ON event_types (is_active);
CREATE INDEX idx_event_types_sort_order ON event_types (sort_order);

-- ================================================
-- STEP 3: Add event_type_id to events table
-- ================================================
ALTER TABLE events ADD COLUMN IF NOT EXISTS event_type_id UUID;

-- ================================================
-- STEP 4: Add foreign key constraint
-- ================================================
ALTER TABLE events ADD CONSTRAINT fk_events_event_type 
    FOREIGN KEY (event_type_id) REFERENCES event_types(id) ON DELETE RESTRICT;

-- ================================================
-- STEP 5: Add index on event_type_id
-- ================================================
CREATE INDEX idx_events_event_type_id ON events (event_type_id);

-- ================================================
-- STEP 6: Create update trigger for updated_at
-- ================================================
CREATE OR REPLACE FUNCTION update_event_types_updated_at() 
RETURNS TRIGGER LANGUAGE plpgsql AS '
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
';

DROP TRIGGER IF EXISTS update_event_types_updated_at ON event_types;
CREATE TRIGGER update_event_types_updated_at
    BEFORE UPDATE ON event_types
    FOR EACH ROW
    EXECUTE FUNCTION update_event_types_updated_at();

-- ================================================
-- STEP 7: Migrate existing data (if any)
-- This maps old 'type' string values to event_type_id
-- ================================================
-- Note: This assumes seeders have been run first
-- If you have existing events with 'type' column, run these after seeding:

-- UPDATE events SET event_type_id = (SELECT id FROM event_types WHERE name = 'workshop') WHERE type = 'workshop';
-- UPDATE events SET event_type_id = (SELECT id FROM event_types WHERE name = 'webinar') WHERE type = 'webinar';
-- UPDATE events SET event_type_id = (SELECT id FROM event_types WHERE name = 'meetup') WHERE type = 'meetup';
-- UPDATE events SET event_type_id = (SELECT id FROM event_types WHERE name = 'bootcamp') WHERE type = 'bootcamp';

-- ================================================
-- STEP 8: Make event_type_id NOT NULL
-- (Only after data migration is complete and verified)
-- ================================================
-- ALTER TABLE events ALTER COLUMN event_type_id SET NOT NULL;

-- ================================================
-- STEP 9: Drop old 'type' column (optional - do after verification)
-- ================================================
-- ALTER TABLE events DROP COLUMN IF EXISTS type;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop trigger
DROP TRIGGER IF EXISTS update_event_types_updated_at ON event_types;
DROP FUNCTION IF EXISTS update_event_types_updated_at();

-- Drop foreign key constraint
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_event_type;

-- Drop index on event_type_id
DROP INDEX IF EXISTS idx_events_event_type_id;

-- Remove event_type_id column from events
ALTER TABLE events DROP COLUMN IF EXISTS event_type_id;

-- Drop event_types table
DROP TABLE IF EXISTS event_types;

-- +goose StatementEnd