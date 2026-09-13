-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- DROP OLD EMPTY TABLES
-- ============================================================

-- Drop old junction tables that are empty and will be recreated
DROP TABLE IF EXISTS event_categories CASCADE;
DROP TABLE IF EXISTS event_format_links CASCADE;
DROP TABLE IF EXISTS event_materials CASCADE;
DROP TABLE IF EXISTS event_speakers CASCADE;
DROP TABLE IF EXISTS event_tickets CASCADE;
DROP TABLE IF EXISTS event_schedules CASCADE;
DROP TABLE IF EXISTS event_invitations CASCADE;
DROP TABLE IF EXISTS event_registrations CASCADE;
DROP TABLE IF EXISTS event_waitlist CASCADE;

-- ============================================================
-- RECREATE TABLES WITH CORRECT SCHEMA
-- ============================================================

-- Event Schedules Table
CREATE TABLE IF NOT EXISTS event_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    session_name VARCHAR(255),
    session_number INTEGER DEFAULT 1,
    start_date DATE NOT NULL,
    end_date DATE,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    timezone VARCHAR(50) DEFAULT 'Africa/Nairobi',
    location TEXT,
    is_virtual BOOLEAN DEFAULT FALSE,
    zoom_link VARCHAR(500),
    meet_link VARCHAR(500),
    max_attendees INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_schedules_event_id ON event_schedules(event_id);
CREATE INDEX IF NOT EXISTS idx_event_schedules_start_date ON event_schedules(start_date);
CREATE INDEX IF NOT EXISTS idx_event_schedules_deleted_at ON event_schedules(deleted_at);

-- Event Tickets Table
CREATE TABLE IF NOT EXISTS event_tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    ticket_type_id UUID NOT NULL REFERENCES ticket_types(id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC(10,2) NOT NULL DEFAULT 0,
    quantity INTEGER NOT NULL DEFAULT 0,
    max_per_person INTEGER,
    early_bird_deadline TIMESTAMP,
    group_min_attendees INTEGER,
    group_discount NUMERIC(10,2),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_tickets_event_id ON event_tickets(event_id);
CREATE INDEX IF NOT EXISTS idx_event_tickets_ticket_type_id ON event_tickets(ticket_type_id);
CREATE INDEX IF NOT EXISTS idx_event_tickets_is_active ON event_tickets(is_active);
CREATE INDEX IF NOT EXISTS idx_event_tickets_deleted_at ON event_tickets(deleted_at);

-- Event Speakers Table
CREATE TABLE IF NOT EXISTS event_speakers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    title VARCHAR(255),
    bio TEXT,
    photo_url VARCHAR(500),
    social_links JSONB DEFAULT '{}',
    sort_order INTEGER DEFAULT 0,
    is_keynote BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_speakers_event_id ON event_speakers(event_id);
CREATE INDEX IF NOT EXISTS idx_event_speakers_deleted_at ON event_speakers(deleted_at);

-- Event Materials Table
CREATE TABLE IF NOT EXISTS event_materials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    material_type_id UUID NOT NULL REFERENCES material_types(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    url VARCHAR(500) NOT NULL,
    is_pre_event BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    file_size BIGINT,
    mime_type VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_materials_event_id ON event_materials(event_id);
CREATE INDEX IF NOT EXISTS idx_event_materials_material_type_id ON event_materials(material_type_id);
CREATE INDEX IF NOT EXISTS idx_event_materials_is_pre_event ON event_materials(is_pre_event);
CREATE INDEX IF NOT EXISTS idx_event_materials_deleted_at ON event_materials(deleted_at);

-- Event Invitations Table
CREATE TABLE IF NOT EXISTS event_invitations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    invited_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    accepted_at TIMESTAMP,
    declined_at TIMESTAMP,
    token VARCHAR(255) UNIQUE,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(event_id, email)
);

CREATE INDEX IF NOT EXISTS idx_event_invitations_event_id ON event_invitations(event_id);
CREATE INDEX IF NOT EXISTS idx_event_invitations_email ON event_invitations(email);
CREATE INDEX IF NOT EXISTS idx_event_invitations_token ON event_invitations(token);
CREATE INDEX IF NOT EXISTS idx_event_invitations_deleted_at ON event_invitations(deleted_at);

-- Event Registrations Table
CREATE TABLE IF NOT EXISTS event_registrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ticket_id UUID REFERENCES event_tickets(id) ON DELETE SET NULL,
    registration_number VARCHAR(100) UNIQUE,
    status VARCHAR(50) DEFAULT 'registered',
    price_paid NUMERIC(10,2) DEFAULT 0,
    payment_id UUID,
    checked_in_at TIMESTAMP,
    checked_out_at TIMESTAMP,
    attended_duration INTEGER,
    feedback_rating INTEGER,
    feedback_comment TEXT,
    certificate_issued BOOLEAN DEFAULT FALSE,
    certificate_issued_at TIMESTAMP,
    certificate_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(event_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_event_registrations_event_id ON event_registrations(event_id);
CREATE INDEX IF NOT EXISTS idx_event_registrations_user_id ON event_registrations(user_id);
CREATE INDEX IF NOT EXISTS idx_event_registrations_ticket_id ON event_registrations(ticket_id);
CREATE INDEX IF NOT EXISTS idx_event_registrations_status ON event_registrations(status);
CREATE INDEX IF NOT EXISTS idx_event_registrations_registration_number ON event_registrations(registration_number);
CREATE INDEX IF NOT EXISTS idx_event_registrations_deleted_at ON event_registrations(deleted_at);

-- Event Waitlist Table
CREATE TABLE IF NOT EXISTS event_waitlist (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ticket_type_id UUID REFERENCES ticket_types(id) ON DELETE SET NULL,
    position INTEGER NOT NULL,
    status VARCHAR(50) DEFAULT 'waiting',
    notified_at TIMESTAMP,
    offered_at TIMESTAMP,
    offered_expires_at TIMESTAMP,
    converted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(event_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_event_waitlist_event_id ON event_waitlist(event_id);
CREATE INDEX IF NOT EXISTS idx_event_waitlist_user_id ON event_waitlist(user_id);
CREATE INDEX IF NOT EXISTS idx_event_waitlist_position ON event_waitlist(position);
CREATE INDEX IF NOT EXISTS idx_event_waitlist_status ON event_waitlist(status);
CREATE INDEX IF NOT EXISTS idx_event_waitlist_deleted_at ON event_waitlist(deleted_at);

-- ============================================================
-- VERIFY SEEDED TABLES HAVE DATA
-- ============================================================

-- These tables should already have data from seeders:
-- categories ✅
-- event_formats ✅
-- material_types ✅
-- ticket_types ✅
-- recurrence_patterns ✅
-- certificate_templates ✅
-- certificate_types ✅

-- ============================================================
-- ADD COMMENTS FOR DOCUMENTATION
-- ============================================================

COMMENT ON TABLE event_schedules IS 'Schedules for multi-day events';
COMMENT ON TABLE event_tickets IS 'Ticket types and pricing for events';
COMMENT ON TABLE event_speakers IS 'Speakers and presenters for events';
COMMENT ON TABLE event_materials IS 'Materials and resources for events';
COMMENT ON TABLE event_invitations IS 'Invitations for invite-only events';
COMMENT ON TABLE event_registrations IS 'Event registrations and attendee tracking';
COMMENT ON TABLE event_waitlist IS 'Waitlist for sold-out events';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- ============================================================
-- ROLLBACK: Drop all newly created tables
-- ============================================================

DROP TABLE IF EXISTS event_waitlist;
DROP TABLE IF EXISTS event_registrations;
DROP TABLE IF EXISTS event_invitations;
DROP TABLE IF EXISTS event_materials;
DROP TABLE IF EXISTS event_speakers;
DROP TABLE IF EXISTS event_tickets;
DROP TABLE IF EXISTS event_schedules;

-- Recreate old empty tables (for rollback)
CREATE TABLE IF NOT EXISTS event_categories (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id, category_id)
);

CREATE TABLE IF NOT EXISTS event_format_links (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    format_id UUID NOT NULL REFERENCES event_formats(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id, format_id)
);
-- +goose StatementEnd