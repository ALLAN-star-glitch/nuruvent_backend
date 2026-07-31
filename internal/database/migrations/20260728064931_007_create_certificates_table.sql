-- +goose Up
CREATE TABLE IF NOT EXISTS certificates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL,
    attendee_id UUID NOT NULL,
    user_id UUID,
    certificate_number VARCHAR(50) UNIQUE NOT NULL,
    template_url VARCHAR(500),
    qr_code VARCHAR(255) UNIQUE,
    is_verified BOOLEAN DEFAULT FALSE,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_certificates_event_id ON certificates(event_id);
CREATE INDEX idx_certificates_attendee_id ON certificates(attendee_id);
CREATE INDEX idx_certificates_user_id ON certificates(user_id);
CREATE INDEX idx_certificates_certificate_number ON certificates(certificate_number);
CREATE INDEX idx_certificates_qr_code ON certificates(qr_code);
CREATE INDEX idx_certificates_deleted_at ON certificates(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS certificates;