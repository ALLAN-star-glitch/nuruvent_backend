-- +goose Up

-- uniq_orders_pending_batch predates the deleted_at-aware version
-- and lacks the soft-delete guard. Both enforce the same rule; keep
-- the stricter one (uniq_orders_active_per_batch_event) and drop the
-- older duplicate.

DROP INDEX IF EXISTS uniq_orders_pending_batch;

-- +goose Down

CREATE UNIQUE INDEX uniq_orders_pending_batch
ON orders ((batch_context->>'event_id'))
WHERE status = 'pending'
  AND kind = 'certificate_batch';