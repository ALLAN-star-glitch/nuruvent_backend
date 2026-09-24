-- +goose Up
-- Replace the unconditional unique index with a partial one that only
-- applies to active registrations. Cancelled/expired rows are excluded,
-- so users can register again for the same event after cancelling.

DROP INDEX IF EXISTS uniq_event_registrations_active;

CREATE UNIQUE INDEX uniq_event_registrations_active
    ON event_registrations (event_id, user_id)
    WHERE is_active = true AND user_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS uniq_event_registrations_active;

CREATE UNIQUE INDEX uniq_event_registrations_active
    ON event_registrations (event_id, user_id)
    WHERE user_id IS NOT NULL;