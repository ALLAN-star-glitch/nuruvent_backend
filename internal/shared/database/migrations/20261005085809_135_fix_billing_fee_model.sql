-- +goose Up
-- +goose StatementBegin

-- The original column name was misleading: this is Nuruvent's cut,
-- not a combined "commission". Rename it and set the correct default
-- (4.5% per the pricing page).
ALTER TABLE orders RENAME COLUMN commission_rate TO platform_fee_rate;
ALTER TABLE orders ALTER COLUMN platform_fee_rate SET DEFAULT 0.0450;
UPDATE orders SET platform_fee_rate = 0.0450;

-- Processing fee depends on the method the attendee picks at checkout
-- (M-Pesa 3.5%, local card 3.5%, international card 4.5%), so it lives
-- on the payment and is snapshotted at initiation. Default of 3.5%
-- matches the common case (M-Pesa and local cards).
ALTER TABLE payments
    ADD COLUMN processing_fee_rate NUMERIC(5,4) NOT NULL DEFAULT 0.0350;

UPDATE payments SET processing_fee_rate = 0.0350;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE payments DROP COLUMN IF EXISTS processing_fee_rate;

ALTER TABLE orders
    ALTER COLUMN platform_fee_rate SET DEFAULT 0.1000;
UPDATE orders SET platform_fee_rate = 0.1000;
ALTER TABLE orders RENAME COLUMN platform_fee_rate TO commission_rate;

-- +goose StatementEnd