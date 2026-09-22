-- +goose Up
-- +goose StatementBegin

DROP TABLE IF EXISTS webhook_events CASCADE;
CREATE TABLE webhook_events (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider           VARCHAR(30)  NOT NULL,
    provider_event_id  VARCHAR(255) NOT NULL,
    signature_valid    BOOLEAN      NOT NULL DEFAULT FALSE,
    payload            BYTEA        NOT NULL,
    received_at        TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at       TIMESTAMPTZ,
    processing_error   TEXT,

    CONSTRAINT chk_webhook_events_provider CHECK (
        provider IN ('mpesa', 'card', 'stub')
    )
);

-- Deduplication key: same provider event cannot be recorded twice.
CREATE UNIQUE INDEX uniq_webhook_events_provider_event
    ON webhook_events (provider, provider_event_id);

CREATE INDEX idx_webhook_events_unprocessed
    ON webhook_events (received_at)
    WHERE processed_at IS NULL OR processing_error IS NOT NULL;

CREATE INDEX idx_webhook_events_received_at
    ON webhook_events (received_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS webhook_events CASCADE;
-- +goose StatementEnd