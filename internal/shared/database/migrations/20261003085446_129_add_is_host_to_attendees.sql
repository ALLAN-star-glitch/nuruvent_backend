-- +goose Up
-- Mark attendees that represent the host of the parent event.
--
-- Host rows are created automatically at publish time (see the
-- events service's attendance_sync). They carry no registration, no
-- ticket, and no payment. The flag lets queries distinguish a host
-- from a real attendee — registered_count excludes hosts, attended_count
-- includes them, and the roster dialog badges them.
ALTER TABLE attendees ADD COLUMN is_host boolean NOT NULL DEFAULT false;

CREATE INDEX idx_attendees_is_host
  ON attendees (is_host)
  WHERE is_host = true;

-- +goose Down
DROP INDEX IF EXISTS idx_attendees_is_host;
ALTER TABLE attendees DROP COLUMN IF EXISTS is_host;