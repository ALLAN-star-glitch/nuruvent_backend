-- +goose Up
-- The uniq_webhook_events_provider_event was created as a UNIQUE
-- INDEX (not a constraint), so DROP CONSTRAINT did not remove it.
--
-- Webhook events are now an append-only audit log. Idempotency is
-- enforced by the payment state machine, not by DB uniqueness.
DROP INDEX IF EXISTS uniq_webhook_events_provider_event;

-- +goose Down
CREATE UNIQUE INDEX uniq_webhook_events_provider_event
    ON webhook_events (provider, provider_event_id);