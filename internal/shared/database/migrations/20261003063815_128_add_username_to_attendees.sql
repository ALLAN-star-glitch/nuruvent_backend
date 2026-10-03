-- +goose Up
-- Denormalized copy of users.username. Written at registration time.
-- Used by the Zoom webhook matcher when Zoom sends a placeholder email
-- and the participant's identity is carried in the name field instead.

ALTER TABLE attendees ADD COLUMN username text;

-- Backfill existing attendees from their linked user, when a link exists.
UPDATE attendees a
SET username = u.username
FROM registrations r
JOIN users u ON u.id = r.user_id
WHERE a.external_type = 'event_registration'
  AND a.external_id = r.id
  AND a.username IS NULL;

CREATE INDEX idx_attendees_username ON attendees (username);

-- +goose Down
DROP INDEX IF EXISTS idx_attendees_username;
ALTER TABLE attendees DROP COLUMN IF EXISTS username;