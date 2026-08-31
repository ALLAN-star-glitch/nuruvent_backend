-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- FIX: events.institution_id foreign key constraint
-- Problem: Currently references users.id but should reference institutions.id
-- ============================================================

-- 1. First, drop the incorrect foreign key constraint
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_institution;

-- 2. Add the correct foreign key constraint referencing institutions
ALTER TABLE events 
    ADD CONSTRAINT fk_events_institution 
    FOREIGN KEY (institution_id) REFERENCES institutions(id) ON DELETE CASCADE;

-- 3. Ensure institution_id is nullable (for personal events)
ALTER TABLE events ALTER COLUMN institution_id DROP NOT NULL;

-- 4. Add an index for performance
CREATE INDEX IF NOT EXISTS idx_events_institution_id_new ON events(institution_id);

-- 5. Update any existing data - set institution_id to NULL for personal events
-- (Optional: Run this if you have data that needs cleaning)
-- UPDATE events SET institution_id = NULL 
-- WHERE institution_id IS NOT NULL 
-- AND NOT EXISTS (SELECT 1 FROM institutions WHERE id = events.institution_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Rollback: Restore the original foreign key constraint

-- Drop the correct foreign key
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_institution;

-- Restore the original foreign key (referencing users)
ALTER TABLE events 
    ADD CONSTRAINT fk_events_institution 
    FOREIGN KEY (institution_id) REFERENCES users(id) ON DELETE CASCADE;

-- Restore NOT NULL constraint if needed
-- ALTER TABLE events ALTER COLUMN institution_id SET NOT NULL;

-- Drop the new index
DROP INDEX IF EXISTS idx_events_institution_id_new;
-- +goose StatementEnd