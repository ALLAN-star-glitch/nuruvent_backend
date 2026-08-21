-- internal/shared/database/migrations/20260820102041_048_fix_event_uuid_nullable.sql

-- +goose Up
-- Make deleted_by and restored_by nullable first (before trying to update)
ALTER TABLE events ALTER COLUMN deleted_by DROP NOT NULL;
ALTER TABLE events ALTER COLUMN restored_by DROP NOT NULL;

-- Remove any default values
ALTER TABLE events ALTER COLUMN deleted_by DROP DEFAULT;
ALTER TABLE events ALTER COLUMN restored_by DROP DEFAULT;

-- +goose Down
ALTER TABLE events ALTER COLUMN deleted_by SET NOT NULL;
ALTER TABLE events ALTER COLUMN restored_by SET NOT NULL;