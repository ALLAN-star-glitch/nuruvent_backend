-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    time VARCHAR(20) NOT NULL,
    duration INTEGER,
    price DECIMAL(10,2) DEFAULT 0,
    certificate_price DECIMAL(10,2) DEFAULT 0,
    location VARCHAR(255),
    is_virtual BOOLEAN DEFAULT TRUE,
    zoom_link VARCHAR(500),
    meet_link VARCHAR(500),
    max_attendees INTEGER,
    current_attendees INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'draft',
    business_id UUID NOT NULL,
    created_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_events_business_id ON events(business_id);
CREATE INDEX idx_events_date ON events(date);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_deleted_at ON events(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS events;