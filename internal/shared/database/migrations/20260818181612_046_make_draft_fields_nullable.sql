-- +goose Up
-- ============================================================
-- Make fields nullable for draft events
-- ============================================================

-- Make event_type_id nullable so drafts can exist without a type
ALTER TABLE events ALTER COLUMN event_type_id DROP NOT NULL;

-- Make date and time nullable for drafts
ALTER TABLE events ALTER COLUMN date DROP NOT NULL;
ALTER TABLE events ALTER COLUMN time DROP NOT NULL;

-- Set defaults for numeric fields
ALTER TABLE events ALTER COLUMN duration SET DEFAULT 0;
ALTER TABLE events ALTER COLUMN max_attendees SET DEFAULT 0;

-- Make location nullable (not required for virtual drafts)
ALTER TABLE events ALTER COLUMN location DROP NOT NULL;

-- Add default for virtual events
ALTER TABLE events ALTER COLUMN is_virtual SET DEFAULT true;

-- Add index for finding drafts without a type (optional but good for performance)
CREATE INDEX idx_events_event_type_id_null ON events(event_type_id) WHERE event_type_id IS NULL;

-- +goose Down
-- ============================================================
-- Rollback changes
-- ============================================================

DROP INDEX IF EXISTS idx_events_event_type_id_null;

ALTER TABLE events ALTER COLUMN event_type_id SET NOT NULL;
ALTER TABLE events ALTER COLUMN date SET NOT NULL;
ALTER TABLE events ALTER COLUMN time SET NOT NULL;
ALTER TABLE events ALTER COLUMN duration DROP DEFAULT;
ALTER TABLE events ALTER COLUMN max_attendees DROP DEFAULT;
ALTER TABLE events ALTER COLUMN location SET NOT NULL;
ALTER TABLE events ALTER COLUMN is_virtual DROP DEFAULT;