-- +goose Up
-- Drop the coarse dedupe constraint.
--
-- Webhook events become an append-only audit log. IntaSend sends
-- multiple webhooks per invoice (PENDING, PROCESSING, COMPLETE) that
-- all share the same provider_event_id — the old unique constraint
-- would silently drop the meaningful ones.
--
-- Idempotency is enforced by the payment state machine
-- (ConfirmPayment/FailPayment), which rejects illegal transitions.
-- The webhook handler returns 200 for already-processed payments.
ALTER TABLE webhook_events
    DROP CONSTRAINT IF EXISTS uniq_webhook_events_provider_event;

-- +goose Down
ALTER TABLE webhook_events
    ADD CONSTRAINT uniq_webhook_events_provider_event
    UNIQUE (provider, provider_event_id);