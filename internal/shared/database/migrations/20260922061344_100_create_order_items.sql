-- +goose Up
-- +goose StatementBegin
CREATE TABLE order_items (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id       UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    ticket_type_id UUID NOT NULL,
    quantity       INTEGER NOT NULL,
    unit_price     BIGINT  NOT NULL,
    discount       BIGINT  NOT NULL DEFAULT 0,
    line_total     BIGINT  NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_order_items_quantity_positive
        CHECK (quantity >= 1),
    CONSTRAINT chk_order_items_unit_price_nonneg
        CHECK (unit_price >= 0),
    CONSTRAINT chk_order_items_discount_nonneg
        CHECK (discount >= 0),
    CONSTRAINT chk_order_items_discount_le_subtotal
        CHECK (discount <= unit_price * quantity),
    CONSTRAINT chk_order_items_line_total_consistent
        CHECK (line_total = unit_price * quantity - discount)
);

CREATE INDEX idx_order_items_order_id       ON order_items (order_id);
CREATE INDEX idx_order_items_ticket_type_id ON order_items (ticket_type_id);

CREATE OR REPLACE FUNCTION update_order_items_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_order_items_updated_at
    BEFORE UPDATE ON order_items
    FOR EACH ROW EXECUTE FUNCTION update_order_items_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_order_items_updated_at ON order_items;
DROP FUNCTION IF EXISTS update_order_items_updated_at();
DROP TABLE IF EXISTS order_items CASCADE;
-- +goose StatementEnd