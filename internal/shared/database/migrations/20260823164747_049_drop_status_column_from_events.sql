-- internal/shared/database/migrations/20260823164747_049_drop_status_column_from_events.sql

-- +goose Up
-- +goose StatementBegin

-- ✅ Drop the legacy status column from events table
-- Using CASCADE to automatically drop dependent objects

-- First, verify all events have an event_status_id
DO $$
DECLARE
    null_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO null_count FROM events WHERE event_status_id IS NULL;
    
    IF null_count > 0 THEN
        RAISE EXCEPTION 'Cannot drop status column: % events have NULL event_status_id. Please migrate data first.', null_count;
    END IF;
END $$;

-- Drop the index
DROP INDEX IF EXISTS idx_events_status;

-- Drop the column with CASCADE to automatically handle dependencies
ALTER TABLE events DROP COLUMN IF EXISTS status CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- ✅ Restore the status column (for rollback)
ALTER TABLE events ADD COLUMN status VARCHAR(50) DEFAULT 'draft';

-- Restore the index
CREATE INDEX idx_events_status ON events(status);

-- Populate the status column from event_status_id
UPDATE events SET status = 'draft' 
WHERE event_status_id = (SELECT id FROM event_statuses WHERE slug = 'event-status-draft');

UPDATE events SET status = 'published' 
WHERE event_status_id = (SELECT id FROM event_statuses WHERE slug = 'event-status-published');

UPDATE events SET status = 'cancelled' 
WHERE event_status_id = (SELECT id FROM event_statuses WHERE slug = 'event-status-cancelled');

UPDATE events SET status = 'completed' 
WHERE event_status_id = (SELECT id FROM event_statuses WHERE slug = 'event-status-completed');

-- +goose StatementEnd