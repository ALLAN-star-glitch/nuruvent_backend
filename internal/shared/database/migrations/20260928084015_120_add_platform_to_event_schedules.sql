-- +goose Up
-- +goose StatementBegin

-- Add platform to event_schedules.
--
-- Stores the video provider chosen for a virtual schedule. One of
-- 'zoom', 'google_meet', or '' when not yet chosen. An empty string
-- means the schedule either has a manually-pasted link or has not
-- yet had a provider selected.
--
-- The video integration reads this in preference to inferring the
-- platform from the join URL, which was unreliable once the URL had
-- been written back to the schedule row.
ALTER TABLE event_schedules
    ADD COLUMN platform VARCHAR(32) NOT NULL DEFAULT '';

-- Partial index for the common query: find virtual schedules that
-- have a platform set. This is used when the events module needs to
-- know which schedules are Nuruvent-managed (as opposed to manually
-- linked).
CREATE INDEX idx_event_schedules_platform
    ON event_schedules (platform)
    WHERE platform <> '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_event_schedules_platform;

ALTER TABLE event_schedules
    DROP COLUMN IF EXISTS platform;

-- +goose StatementEnd