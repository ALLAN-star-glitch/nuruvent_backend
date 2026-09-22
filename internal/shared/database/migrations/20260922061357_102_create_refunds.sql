-- +goose Up
-- +goose StatementBegin

DROP TABLE IF EXISTS refunds CASCADE;
CREATE TABLE refunds (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_id          UUID NOT NULL REFERENCES payments(id) ON DELETE RESTRICT,
    amount              BIGINT       NOT NULL,
    currency            VARCHAR(3)   NOT NULL DEFAULT 'KES',
    reason              TEXT,
    actor_id            UUID         NOT NULL,
    provider_reference  VARCHAR(255),
    status              VARCHAR(20)  NOT NULL DEFAULT 'pending',
    failure_reason      TEXT,
    completed_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_refunds_amount_positive CHECK (amount > 0),
    CONSTRAINT chk_refunds_status CHECK (
        status IN ('pending', 'succeeded', 'failed')
    ),
    CONSTRAINT chk_refunds_currency CHECK (
        currency IN ('KES', 'USD')
    )
);

CREATE INDEX idx_refunds_payment_id    ON refunds (payment_id);
CREATE INDEX idx_refunds_status        ON refunds (status);
CREATE INDEX idx_refunds_actor_id      ON refunds (actor_id);
CREATE INDEX idx_refunds_created_at    ON refunds (created_at DESC);

CREATE OR REPLACE FUNCTION update_refunds_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_refunds_updated_at
    BEFORE UPDATE ON refunds
    FOR EACH ROW EXECUTE FUNCTION update_refunds_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_refunds_updated_at ON refunds;
DROP FUNCTION IF EXISTS update_refunds_updated_at();
DROP TABLE IF EXISTS refunds CASCADE;
-- +goose StatementEnd