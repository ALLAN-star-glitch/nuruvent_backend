-- +goose Up

-- Host attendee rows use external_type = 'event_host', not 'event'.
-- The previous backfill (20261003213506) filtered on 'event' and
-- therefore matched no host rows. This corrects that.
--
-- Idempotent: guarded by a.phone = '' so re-runs touch nothing.

UPDATE attendees a
SET phone = acc.phone,
    email = CASE WHEN a.email = '' THEN acc.email ELSE a.email END
FROM events e
JOIN teams t ON t.id = e.team_id
JOIN accounts acc ON acc.id = t.account_id
WHERE a.is_host = true
  AND a.external_type = 'event_host'
  AND a.external_id = e.id
  AND a.phone = ''
  AND acc.phone IS NOT NULL
  AND acc.phone <> '';

-- +goose Down
SELECT 1;