-- +goose Up

-- event_schedules.video_meeting_id currently holds the platform-side
-- meeting code (e.g. "spaces/ot0nSSlGgp4B") rather than the Nuruvent
-- UUID it was named for. This column adds the correct semantics:
-- video_meeting_external_id holds the platform code, video_meeting_id
-- (when set going forward) holds video_meetings.id.
--
-- Existing rows keep their current value in video_meeting_id until
-- they are regenerated or updated. Backfill is intentionally not
-- attempted: the two tables were never reliably linked, so any
-- automatic backfill would be guesswork.

ALTER TABLE event_schedules
    ADD COLUMN IF NOT EXISTS video_meeting_external_id VARCHAR(255);

-- Index it because the attendance sync reads this column per session
-- on every event summary request.
CREATE INDEX IF NOT EXISTS idx_event_schedules_video_meeting_external_id
    ON event_schedules (video_meeting_external_id)
    WHERE video_meeting_external_id IS NOT NULL;

-- +goose Down

DROP INDEX IF EXISTS idx_event_schedules_video_meeting_external_id;

ALTER TABLE event_schedules
    DROP COLUMN IF EXISTS video_meeting_external_id;