-- +goose Up

-- ============================================================
-- event_schedules.video_meeting_id — semantic correction
-- ============================================================
--
-- The column was previously constrained to hold a platform-side
-- meeting code (e.g. "spaces/ot0nSSlGgp4B"), not a Nuruvent UUID.
-- That decision is the root of the attendance-orphan bug: the
-- attendance sync parses event_schedules.meet_link to produce a code,
-- while video_meetings stores the code in external_id — and the two
-- sides rarely matched.
--
-- The correct semantics are:
--
--   video_meeting_id            → FK to video_meetings.id (uuid)
--   video_meeting_external_id   → platform code (spaces/..., Zoom ID)
--
-- This migration:
--   1. Adds the new external-id column.
--   2. Drops the check constraint that enforced the old (wrong)
--      shape on video_meeting_id.
--   3. Changes video_meeting_id's type from text to uuid.
--   4. Adds the FK to video_meetings.

-- ------------------------------------------------------------
-- 1. New column for the platform-side code.
-- ------------------------------------------------------------

ALTER TABLE event_schedules
    ADD COLUMN IF NOT EXISTS video_meeting_external_id VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_event_schedules_video_meeting_external_id
    ON event_schedules (video_meeting_external_id)
    WHERE video_meeting_external_id IS NOT NULL;

-- ------------------------------------------------------------
-- 2. Drop the constraint that pinned the old semantic.
-- ------------------------------------------------------------
--
-- video_meeting_id_not_uuid enforced: "value must NOT be a UUID."
-- The column is now the FK, so it must be a UUID.

ALTER TABLE event_schedules
    DROP CONSTRAINT IF EXISTS video_meeting_id_not_uuid;

-- ------------------------------------------------------------
-- 3. Coerce legacy values before the type change.
-- ------------------------------------------------------------
--
-- Any surviving value that is neither NULL nor a UUID is a
-- platform code (legacy). Move it to the new column and clear
-- the FK so the cast succeeds. Rows fixed this way will be
-- re-linked on the next regenerate or publish.

UPDATE event_schedules
SET video_meeting_external_id = video_meeting_id,
    video_meeting_id = NULL
WHERE video_meeting_id IS NOT NULL
  AND video_meeting_id !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

-- ------------------------------------------------------------
-- 4. Type change.
-- ------------------------------------------------------------

ALTER TABLE event_schedules
    ALTER COLUMN video_meeting_id TYPE uuid
    USING video_meeting_id::uuid;

-- ------------------------------------------------------------
-- 5. FK to video_meetings.
-- ------------------------------------------------------------
--
-- ON DELETE SET NULL: when a meeting is deleted (e.g. during a
-- regenerate), the schedule should lose its link, not the whole
-- row.

ALTER TABLE event_schedules
    ADD CONSTRAINT event_schedules_video_meeting_id_fkey
    FOREIGN KEY (video_meeting_id)
    REFERENCES video_meetings(id)
    ON DELETE SET NULL;

-- +goose Down

-- Drop the FK.
ALTER TABLE event_schedules
    DROP CONSTRAINT IF EXISTS event_schedules_video_meeting_id_fkey;

-- Revert the type. Data loss is not recoverable here (UUIDs would
-- round-trip as text, but the original meaning is gone), so this
-- is best-effort only.
ALTER TABLE event_schedules
    ALTER COLUMN video_meeting_id TYPE text
    USING video_meeting_id::text;

-- The video_meeting_id_not_uuid check constraint is intentionally
-- not restored. It enforced the pre-fix behavior and would reject
-- valid post-fix data.

-- Drop the new column and its index.
DROP INDEX IF EXISTS idx_event_schedules_video_meeting_external_id;

ALTER TABLE event_schedules
    DROP COLUMN IF EXISTS video_meeting_external_id;