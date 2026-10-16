-- +goose Up

-- The previous unique index enforced "one pending order per
-- registration" but also silently blocked a second pending
-- certificate_batch order, because batch orders carry an empty
-- registration_id and the index couldn't tell them apart.
--
-- Scope the registration constraint to registration orders, and add
-- a separate constraint for batch orders keyed by event.

DROP INDEX IF EXISTS uniq_orders_active_per_registration;

CREATE UNIQUE INDEX uniq_orders_active_per_registration
ON orders (registration_id)
WHERE status = 'pending'
  AND kind = 'registration'
  AND deleted_at IS NULL;

CREATE UNIQUE INDEX uniq_orders_active_per_batch_event
ON orders ((batch_context->>'event_id'))
WHERE status = 'pending'
  AND kind = 'certificate_batch'
  AND deleted_at IS NULL;

-- +goose Down

DROP INDEX IF EXISTS uniq_orders_active_per_batch_event;

DROP INDEX IF EXISTS uniq_orders_active_per_registration;

CREATE UNIQUE INDEX uniq_orders_active_per_registration
ON orders (registration_id)
WHERE status = 'pending'
  AND deleted_at IS NULL;