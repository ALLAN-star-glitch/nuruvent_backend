-- +goose Up
-- +goose StatementBegin

-- Derived event-level fields. Populated by the events service from
-- schedules. The client no longer supplies these; they are read-only
-- from the API consumer's point of view.
ALTER TABLE events
    ADD COLUMN IF NOT EXISTS date      TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS time      VARCHAR(8) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS duration  INTEGER    NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_events_date ON events (date);

-- Video meeting reference on the schedule. Set when the events service
-- creates a meeting through the video module on behalf of the host.
ALTER TABLE event_schedules
    ADD COLUMN IF NOT EXISTS video_meeting_id UUID;

CREATE INDEX IF NOT EXISTS idx_event_schedules_video_meeting_id
    ON event_schedules (video_meeting_id)
    WHERE video_meeting_id IS NOT NULL;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_event_schedules_video_meeting_id;
ALTER TABLE event_schedules DROP COLUMN IF EXISTS video_meeting_id;

DROP INDEX IF EXISTS idx_events_date;
ALTER TABLE events
    DROP COLUMN IF EXISTS date,
    DROP COLUMN IF EXISTS time,
    DROP COLUMN IF EXISTS duration;
-- +goose StatementEnd