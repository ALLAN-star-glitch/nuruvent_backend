-- +goose Up
-- +goose StatementBegin

-- processing_fee_rate snapshots the payment provider's cut at the moment
-- the payment method is known (M-Pesa 1.5%, local card 2.9%, intl card 3.8%).
-- Mirrors platform_fee_rate's type and constraints exactly.
-- DEFAULT 0 keeps all historical orders unchanged — their NetToOrganizer
-- math remains exactly as it was before this migration. Only new orders
-- written after this migration will carry a non-zero rate.
ALTER TABLE orders
    ADD COLUMN processing_fee_rate numeric(5,4) NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE orders
    DROP COLUMN IF EXISTS processing_fee_rate;
-- +goose StatementEnd