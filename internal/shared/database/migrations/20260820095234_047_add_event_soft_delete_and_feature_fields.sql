-- internal/shared/database/migrations/XXX_add_event_soft_delete_and_feature_fields.sql

-- +goose Up
-- ============================================================
-- Add soft delete tracking fields
-- ============================================================

-- Add deleted_by column to track who deleted the event
ALTER TABLE events ADD COLUMN deleted_by UUID NULL;

-- Add restored_at column to track when event was restored
ALTER TABLE events ADD COLUMN restored_at TIMESTAMP NULL;

-- Add restored_by column to track who restored the event
ALTER TABLE events ADD COLUMN restored_by UUID NULL;

-- ============================================================
-- Add feature flag fields
-- ============================================================

-- Add is_featured column for featured events
ALTER TABLE events ADD COLUMN is_featured BOOLEAN DEFAULT false;

-- Add is_private column for private events
ALTER TABLE events ADD COLUMN is_private BOOLEAN DEFAULT false;

-- ============================================================
-- Add indexes for performance
-- ============================================================

-- Index for deleted_by lookups
CREATE INDEX idx_events_deleted_by ON events(deleted_by);

-- Index for restored_at lookups
CREATE INDEX idx_events_restored_at ON events(restored_at);

-- Index for restored_by lookups
CREATE INDEX idx_events_restored_by ON events(restored_by);

-- Index for is_featured lookups
CREATE INDEX idx_events_is_featured ON events(is_featured);

-- Index for is_private lookups
CREATE INDEX idx_events_is_private ON events(is_private);

-- ============================================================
-- Update existing records to have default values
-- ============================================================

UPDATE events SET 
    is_featured = false,
    is_private = false
WHERE is_featured IS NULL OR is_private IS NULL;

-- ============================================================
-- Create views for common queries
-- ============================================================

-- View for active (non-deleted) events
CREATE OR REPLACE VIEW active_events AS
SELECT * FROM events WHERE deleted_at IS NULL;

-- View for public events (non-deleted and not private)
CREATE OR REPLACE VIEW public_events AS
SELECT * FROM events WHERE deleted_at IS NULL AND is_private = false;

-- ============================================================
-- Comments for documentation
-- ============================================================

COMMENT ON COLUMN events.deleted_by IS 'User ID who soft-deleted the event';
COMMENT ON COLUMN events.restored_at IS 'Timestamp when the event was restored';
COMMENT ON COLUMN events.restored_by IS 'User ID who restored the event';
COMMENT ON COLUMN events.is_featured IS 'Whether the event is featured on the platform';
COMMENT ON COLUMN events.is_private IS 'Whether the event is private (only visible to creator/team)';

-- +goose Down
-- ============================================================
-- Rollback changes
-- ============================================================

-- Drop views
DROP VIEW IF EXISTS public_events;
DROP VIEW IF EXISTS active_events;

-- Drop indexes
DROP INDEX IF EXISTS idx_events_deleted_by;
DROP INDEX IF EXISTS idx_events_restored_at;
DROP INDEX IF EXISTS idx_events_restored_by;
DROP INDEX IF EXISTS idx_events_is_featured;
DROP INDEX IF EXISTS idx_events_is_private;

-- Drop columns
ALTER TABLE events DROP COLUMN IF EXISTS deleted_by;
ALTER TABLE events DROP COLUMN IF EXISTS restored_at;
ALTER TABLE events DROP COLUMN IF EXISTS restored_by;
ALTER TABLE events DROP COLUMN IF EXISTS is_featured;
ALTER TABLE events DROP COLUMN IF EXISTS is_private;

-- Drop comments
COMMENT ON COLUMN events.deleted_by IS NULL;
COMMENT ON COLUMN events.restored_at IS NULL;
COMMENT ON COLUMN events.restored_by IS NULL;
COMMENT ON COLUMN events.is_featured IS NULL;
COMMENT ON COLUMN events.is_private IS NULL;