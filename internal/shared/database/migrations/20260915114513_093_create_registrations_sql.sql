-- +goose Up
-- +goose StatementBegin
CREATE TABLE registrations (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID REFERENCES users(id) ON DELETE SET NULL,
    guest_email         VARCHAR(255),
    guest_name          VARCHAR(255),
    guest_phone         VARCHAR(20),
    registration_number VARCHAR(100) NOT NULL UNIQUE,
    status_id           UUID NOT NULL REFERENCES registration_statuses(id),
    currency            VARCHAR(3) NOT NULL DEFAULT 'KES',
    subtotal            BIGINT NOT NULL DEFAULT 0,
    discount_total      BIGINT NOT NULL DEFAULT 0,
    total_amount        BIGINT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    confirmed_at        TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,
    cancelled_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    cancellation_reason TEXT,
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT chk_registrations_identity CHECK (
        (user_id IS NOT NULL AND guest_email IS NULL) OR
        (user_id IS NULL AND guest_email IS NOT NULL)
    ),
    CONSTRAINT chk_registrations_amounts CHECK (
        subtotal >= 0
        AND discount_total >= 0
        AND total_amount >= 0
        AND discount_total <= subtotal
        AND total_amount = subtotal - discount_total
    )
);

CREATE INDEX idx_registrations_user_id     ON registrations(user_id);
CREATE INDEX idx_registrations_status_id   ON registrations(status_id);
CREATE INDEX idx_registrations_guest_email ON registrations(guest_email);
CREATE INDEX idx_registrations_deleted_at  ON registrations(deleted_at);

CREATE OR REPLACE FUNCTION update_registrations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_registrations_updated_at
    BEFORE UPDATE ON registrations
    FOR EACH ROW EXECUTE FUNCTION update_registrations_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_registrations_updated_at ON registrations;
DROP FUNCTION IF EXISTS update_registrations_updated_at();
DROP TABLE IF EXISTS registrations CASCADE;
-- +goose StatementEnd