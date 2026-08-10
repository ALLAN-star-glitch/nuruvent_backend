-- +goose Up
-- +goose StatementBegin

-- ================================================
-- STEP 1: Create event_statuses table
-- ================================================
CREATE TABLE IF NOT EXISTS event_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(20),
    icon VARCHAR(50),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    is_final BOOLEAN DEFAULT false,
    can_edit BOOLEAN DEFAULT true,
    can_register BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- ================================================
-- STEP 2: Add indexes
-- ================================================
CREATE INDEX idx_event_statuses_name ON event_statuses (name);
CREATE INDEX idx_event_statuses_is_active ON event_statuses (is_active);
CREATE INDEX idx_event_statuses_sort_order ON event_statuses (sort_order);

-- ================================================
-- STEP 3: Add event_status_id to events table
-- ================================================
ALTER TABLE events ADD COLUMN IF NOT EXISTS event_status_id UUID;

-- ================================================
-- STEP 4: Add foreign key constraint
-- ================================================
ALTER TABLE events ADD CONSTRAINT fk_events_event_status 
    FOREIGN KEY (event_status_id) REFERENCES event_statuses(id) ON DELETE RESTRICT;

-- ================================================
-- STEP 5: Add index on event_status_id
-- ================================================
CREATE INDEX idx_events_event_status_id ON events (event_status_id);

-- ================================================
-- STEP 6: Create update trigger for updated_at
-- ================================================
CREATE OR REPLACE FUNCTION update_event_statuses_updated_at() 
RETURNS TRIGGER LANGUAGE plpgsql AS '
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
';

DROP TRIGGER IF EXISTS update_event_statuses_updated_at ON event_statuses;
CREATE TRIGGER update_event_statuses_updated_at
    BEFORE UPDATE ON event_statuses
    FOR EACH ROW
    EXECUTE FUNCTION update_event_statuses_updated_at();

-- ================================================
-- STEP 7: Migrate existing data (if any)
-- This maps old 'status' string values to event_status_id
-- Note: This assumes seeders have been run first
-- ================================================

-- Update events based on their status
-- UPDATE events SET event_status_id = (SELECT id FROM event_statuses WHERE name = 'draft') WHERE status = 'draft';
-- UPDATE events SET event_status_id = (SELECT id FROM event_statuses WHERE name = 'published') WHERE status = 'published';
-- UPDATE events SET event_status_id = (SELECT id FROM event_statuses WHERE name = 'cancelled') WHERE status = 'cancelled';
-- UPDATE events SET event_status_id = (SELECT id FROM event_statuses WHERE name = 'completed') WHERE status = 'completed';

-- ================================================
-- STEP 8: Make event_status_id NOT NULL (after verification)
-- ================================================
-- ALTER TABLE events ALTER COLUMN event_status_id SET NOT NULL;

-- ================================================
-- STEP 9: Drop old 'status' column (after verification)
-- ================================================
-- ALTER TABLE events DROP COLUMN IF EXISTS status;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop trigger
DROP TRIGGER IF EXISTS update_event_statuses_updated_at ON event_statuses;
DROP FUNCTION IF EXISTS update_event_statuses_updated_at();

-- Drop foreign key
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_event_status;

-- Drop index
DROP INDEX IF EXISTS idx_events_event_status_id;

-- Remove column
ALTER TABLE events DROP COLUMN IF EXISTS event_status_id;

-- Drop table
DROP TABLE IF EXISTS event_statuses;

-- +goose StatementEnd