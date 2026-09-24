-- +goose Up
-- +goose StatementBegin

-- The legacy `payments` table was created by an early migration
-- (008_create_payments_table.sql) as a placeholder. This module's
-- schema replaces it. Drop the placeholder before creating the real
-- schema so the migration is idempotent and works on any DB state.
DROP TABLE IF EXISTS payments CASCADE;

CREATE TABLE payments (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id            UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    provider            VARCHAR(30)  NOT NULL,
    method              VARCHAR(20)  NOT NULL,
    amount              BIGINT       NOT NULL,
    currency            VARCHAR(3)   NOT NULL DEFAULT 'KES',
    status              VARCHAR(20)  NOT NULL DEFAULT 'pending',
    idempotency_key     VARCHAR(100) NOT NULL,
    provider_reference  VARCHAR(255),
    failure_reason      TEXT,
    initiated_at        TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at          TIMESTAMPTZ  NOT NULL,
    completed_at        TIMESTAMPTZ,
    failed_at           TIMESTAMPTZ,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_payments_amount_positive CHECK (amount > 0),
    CONSTRAINT chk_payments_method CHECK (
        method IN ('mpesa', 'card')
    ),
    CONSTRAINT chk_payments_status CHECK (
        status IN ('pending', 'succeeded', 'failed', 'expired', 'refunded')
    ),
    CONSTRAINT chk_payments_currency CHECK (
        currency IN ('KES', 'USD')
    ),
    CONSTRAINT chk_payments_succeeded_requires_reference CHECK (
        status != 'succeeded' OR provider_reference IS NOT NULL
    )
);

CREATE UNIQUE INDEX uniq_payments_order_idempotency_key
    ON payments (order_id, idempotency_key);

CREATE UNIQUE INDEX uniq_payments_provider_reference
    ON payments (provider, provider_reference)
    WHERE provider_reference IS NOT NULL;

CREATE INDEX idx_payments_order_id  ON payments (order_id);
CREATE INDEX idx_payments_status    ON payments (status);
CREATE INDEX idx_payments_expires_at ON payments (expires_at)
    WHERE status = 'pending';

CREATE OR REPLACE FUNCTION update_payments_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_payments_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP FUNCTION IF EXISTS update_payments_updated_at();
DROP TABLE IF EXISTS payments CASCADE;
-- +goose StatementEnd