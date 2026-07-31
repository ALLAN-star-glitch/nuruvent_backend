-- +goose Up
CREATE TABLE IF NOT EXISTS payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id UUID NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'KES',
    method VARCHAR(50) DEFAULT 'mpesa',
    reference VARCHAR(255) UNIQUE,
    status VARCHAR(50) DEFAULT 'pending',
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_payouts_business_id ON payouts(business_id);
CREATE INDEX idx_payouts_reference ON payouts(reference);
CREATE INDEX idx_payouts_status ON payouts(status);
CREATE INDEX idx_payouts_deleted_at ON payouts(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS payouts;