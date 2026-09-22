-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    registration_id UUID NOT NULL,
    user_id         UUID,
    guest_email     VARCHAR(255),
    currency        VARCHAR(3)   NOT NULL DEFAULT 'KES',
    subtotal        BIGINT       NOT NULL DEFAULT 0,
    discount_total  BIGINT       NOT NULL DEFAULT 0,
    total_amount    BIGINT       NOT NULL DEFAULT 0,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
    expires_at      TIMESTAMPTZ  NOT NULL,
    paid_at         TIMESTAMPTZ,
    cancelled_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT chk_orders_identity CHECK (
        (user_id IS NOT NULL AND guest_email IS NULL) OR
        (user_id IS NULL AND guest_email IS NOT NULL)
    ),
    CONSTRAINT chk_orders_amounts CHECK (
        subtotal      >= 0
        AND discount_total >= 0
        AND total_amount   >= 0
        AND discount_total <= subtotal
        AND total_amount = subtotal - discount_total
    ),
    CONSTRAINT chk_orders_status CHECK (
        status IN ('pending', 'paid', 'expired', 'cancelled')
    ),
    CONSTRAINT chk_orders_currency CHECK (
        currency IN ('KES', 'USD')
    )
);

-- At most one pending order per registration. This enforces the
-- one-active-order-per-registration rule at the DB level.
CREATE UNIQUE INDEX uniq_orders_active_per_registration
    ON orders (registration_id)
    WHERE status = 'pending' AND deleted_at IS NULL;

CREATE INDEX idx_orders_registration_id ON orders (registration_id);
CREATE INDEX idx_orders_user_id         ON orders (user_id);
CREATE INDEX idx_orders_status          ON orders (status);
CREATE INDEX idx_orders_expires_at      ON orders (expires_at)
    WHERE status = 'pending';
CREATE INDEX idx_orders_deleted_at      ON orders (deleted_at);

CREATE OR REPLACE FUNCTION update_orders_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_orders_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP FUNCTION IF EXISTS update_orders_updated_at();
DROP TABLE IF EXISTS orders CASCADE;
-- +goose StatementEnd