-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- MIGRATION: Rename account_id to institution_id in events table
-- Description: Aligns events table with new auth structure where
--              events belong to institutions (not accounts)
-- ============================================================

-- ============================================================
-- 1. RENAME COLUMN
-- ============================================================
ALTER TABLE events RENAME COLUMN account_id TO institution_id;

-- ============================================================
-- 2. RENAME INDEX
-- ============================================================
ALTER INDEX IF EXISTS idx_events_account_id RENAME TO idx_events_institution_id;

-- ============================================================
-- 3. UPDATE FOREIGN KEY CONSTRAINT (if it exists)
-- ============================================================
-- PostgreSQL doesn't support IF EXISTS with RENAME CONSTRAINT
-- Check if constraint exists before renaming
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'fk_events_account' 
        AND conrelid = 'events'::regclass
    ) THEN
        EXECUTE 'ALTER TABLE events RENAME CONSTRAINT fk_events_account TO fk_events_institution';
    END IF;
END $$;

-- ============================================================
-- 4. UPDATE COMMENT
-- ============================================================
COMMENT ON COLUMN events.institution_id IS 'The institution this event belongs to';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- ============================================================
-- DOWN: Revert changes
-- ============================================================

ALTER TABLE events RENAME COLUMN institution_id TO account_id;

ALTER INDEX IF EXISTS idx_events_institution_id RENAME TO idx_events_account_id;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'fk_events_institution' 
        AND conrelid = 'events'::regclass
    ) THEN
        EXECUTE 'ALTER TABLE events RENAME CONSTRAINT fk_events_institution TO fk_events_account';
    END IF;
END $$;

COMMENT ON COLUMN events.account_id IS 'The account this event belongs to';

-- +goose StatementEnd