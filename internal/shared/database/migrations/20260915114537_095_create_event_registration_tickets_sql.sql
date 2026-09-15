-- +goose Up
-- +goose StatementBegin
CREATE TABLE event_registration_tickets (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    registration_id UUID NOT NULL REFERENCES registrations(id) ON DELETE CASCADE,
    ticket_type_id  UUID NOT NULL REFERENCES ticket_types(id),
    quantity        INTEGER NOT NULL,
    unit_price      BIGINT NOT NULL,
    discount        BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_ert_quantity_positive CHECK (quantity > 0),
    CONSTRAINT chk_ert_unit_price_nonneg CHECK (unit_price >= 0),
    CONSTRAINT chk_ert_discount_nonneg   CHECK (discount >= 0),
    CONSTRAINT chk_ert_discount_le_subtotal
        CHECK (discount <= unit_price * quantity)
);

CREATE INDEX idx_ert_registration_id ON event_registration_tickets(registration_id);
CREATE INDEX idx_ert_ticket_type_id  ON event_registration_tickets(ticket_type_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS event_registration_tickets CASCADE;
-- +goose StatementEnd