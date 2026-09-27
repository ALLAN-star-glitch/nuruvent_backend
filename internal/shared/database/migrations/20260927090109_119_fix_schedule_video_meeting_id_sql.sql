-- +goose Up

-- event_schedules.video_meeting_id was declared as UUID, but the
-- update and delete paths need to store Zoom's external meeting ID,
-- which is a numeric string (e.g. "74911910997") — not a UUID.
--
-- Step 1: alter the column type from UUID to TEXT. Existing UUIDs
-- are preserved as their text representations.
--
-- Step 2: backfill every non-null value with the corresponding
-- external ID from video_meetings. Both sides are cast to text so the
-- comparison works after the column change.

ALTER TABLE event_schedules
  ALTER COLUMN video_meeting_id TYPE TEXT USING video_meeting_id::text;

UPDATE event_schedules s
SET video_meeting_id = m.external_id
FROM video_meetings m
WHERE m.id::text = s.video_meeting_id::text
  AND s.video_meeting_id IS NOT NULL;


-- Step 3: reject UUID-shaped values going forward. Zoom external IDs
-- are numeric, so a UUID in this column is a bug.

ALTER TABLE event_schedules
  ADD CONSTRAINT video_meeting_id_not_uuid
  CHECK (
    video_meeting_id IS NULL
    OR video_meeting_id !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
  );


-- +goose Down

ALTER TABLE event_schedules
  DROP CONSTRAINT IF EXISTS video_meeting_id_not_uuid;

-- Convert back to UUID. Rows whose current value is not a valid UUID
-- (i.e. the migrated external IDs) are set to NULL first — UUID typing
-- cannot hold numeric strings.

UPDATE event_schedules
SET video_meeting_id = NULL
WHERE video_meeting_id IS NOT NULL
  AND video_meeting_id !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

ALTER TABLE event_schedules
  ALTER COLUMN video_meeting_id TYPE UUID USING video_meeting_id::uuid;