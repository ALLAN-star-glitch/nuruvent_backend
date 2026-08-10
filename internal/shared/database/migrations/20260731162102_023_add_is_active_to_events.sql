-- +goose Up
-- +goose StatementBegin

-- Add is_active column to events table
ALTER TABLE events ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_events_is_active ON events (is_active);

-- Update existing events to be active by default
UPDATE events SET is_active = true WHERE is_active IS NULL;

-- Add comment for documentation
COMMENT ON COLUMN events.is_active IS 'Business logic flag to manually deactivate events without soft deleting';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop index
DROP INDEX IF EXISTS idx_events_is_active;

-- Drop column
ALTER TABLE events DROP COLUMN IF EXISTS is_active;

-- +goose StatementEnd