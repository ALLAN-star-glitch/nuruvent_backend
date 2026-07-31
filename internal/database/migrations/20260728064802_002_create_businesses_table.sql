-- +goose Up
CREATE TABLE IF NOT EXISTS businesses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    business_type_id UUID,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(20),
    address TEXT,
    logo VARCHAR(500),
    website VARCHAR(255),
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    is_email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP,
    is_verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP,
    verified_by UUID,
    verification_method VARCHAR(50),
    created_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_businesses_name ON businesses(name);
CREATE INDEX idx_businesses_email ON businesses(email);
CREATE INDEX idx_businesses_is_active ON businesses(is_active);
CREATE INDEX idx_businesses_deleted_at ON businesses(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS businesses;