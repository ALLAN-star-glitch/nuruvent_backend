-- +goose Up
-- Drop the stale provider CHECK constraint on webhook_events.
-- It only allowed 'mpesa', 'card', and 'stub' — but the architecture
-- now treats `provider` as a provider name ('intasend', 'flutterwave',
-- etc.), which are registered dynamically at runtime.
ALTER TABLE webhook_events DROP CONSTRAINT IF EXISTS chk_webhook_events_provider;

-- +goose Down
-- Restore the original constraint for rollback.
ALTER TABLE webhook_events
ADD CONSTRAINT chk_webhook_events_provider
CHECK (provider IN ('mpesa', 'card', 'stub'));