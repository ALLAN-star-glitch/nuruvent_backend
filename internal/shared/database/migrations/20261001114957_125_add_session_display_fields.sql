-- +goose Up

-- Session display fields.
--
-- Denormalized from the parent event at sync time so the join-token
-- redemption path can build the frontend redirect URL (/meeting/:code
-- with name, host, return, platform query params) without a
-- cross-module lookup during the public join flow.
--
-- Populated by the events module's syncEventSchedulesToAttendance,
-- which runs on every event update. If the event's display name or
-- organizer changes later, the next sync refreshes these columns.
ALTER TABLE sessions ADD COLUMN event_display_name     text NOT NULL DEFAULT '';
ALTER TABLE sessions ADD COLUMN organizer_display_name text NOT NULL DEFAULT '';

-- +goose Down

ALTER TABLE sessions DROP COLUMN IF EXISTS event_display_name;
ALTER TABLE sessions DROP COLUMN IF EXISTS organizer_display_name;