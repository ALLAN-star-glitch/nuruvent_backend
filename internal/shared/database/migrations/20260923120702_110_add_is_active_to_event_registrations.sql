-- +goose Up
-- Add is_active column. Controls whether this registration counts as
-- "live" for the partial unique index. Cancelled and expired
-- registrations flip to false so the user can re-register.

ALTER TABLE event_registrations
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- Backfill: mark already-cancelled/expired registrations as inactive.
UPDATE event_registrations er
SET is_active = false
FROM registrations r
JOIN registration_statuses rs ON rs.id = r.status_id
WHERE r.id = er.registration_id
  AND rs.slug IN ('cancelled', 'expired');

-- +goose Down
ALTER TABLE event_registrations DROP COLUMN is_active;