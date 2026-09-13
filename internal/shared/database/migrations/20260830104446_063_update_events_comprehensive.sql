-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- ADD COMPREHENSIVE EVENT FIELDS
-- ============================================================

-- 1. Basic Information Fields
ALTER TABLE events 
    ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS short_description TEXT,
    ADD COLUMN IF NOT EXISTS language VARCHAR(10) DEFAULT 'en',
    ADD COLUMN IF NOT EXISTS start_date TIMESTAMP,
    ADD COLUMN IF NOT EXISTS end_date TIMESTAMP,
    ADD COLUMN IF NOT EXISTS is_multi_day BOOLEAN DEFAULT FALSE;

-- 2. Recurrence Fields (already partially added in previous migration)
-- Adding remaining recurrence fields
ALTER TABLE events 
    ADD COLUMN IF NOT EXISTS recurrence_days_of_week TEXT[],
    ADD COLUMN IF NOT EXISTS recurrence_day_of_month INTEGER,
    ADD COLUMN IF NOT EXISTS recurrence_week_of_month VARCHAR(20);

-- 3. Venue & Location Enhancements
ALTER TABLE events 
    ADD COLUMN IF NOT EXISTS venue_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS venue_address TEXT,
    ADD COLUMN IF NOT EXISTS venue_city VARCHAR(100),
    ADD COLUMN IF NOT EXISTS venue_country VARCHAR(100),
    ADD COLUMN IF NOT EXISTS venue_coordinates JSONB,
    ADD COLUMN IF NOT EXISTS is_hybrid BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS virtual_platform VARCHAR(50),
    ADD COLUMN IF NOT EXISTS virtual_platform_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS in_person_location TEXT;

-- 4. Ticketing & Capacity Enhancements
ALTER TABLE events 
    ADD COLUMN IF NOT EXISTS waitlist_capacity INTEGER,
    ADD COLUMN IF NOT EXISTS ticket_sales_start TIMESTAMP,
    ADD COLUMN IF NOT EXISTS ticket_sales_end TIMESTAMP,
    ADD COLUMN IF NOT EXISTS min_tickets_per_order INTEGER DEFAULT 1,
    ADD COLUMN IF NOT EXISTS max_tickets_per_order INTEGER;

-- 5. Access & Privacy Enhancements
ALTER TABLE events 
    ADD COLUMN IF NOT EXISTS invited_emails TEXT[] DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS requires_approval BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS approval_required_for TEXT[];

-- 6. Monetization & Add-ons
ALTER TABLE events 
    ADD COLUMN IF NOT EXISTS featured_until TIMESTAMP,
    ADD COLUMN IF NOT EXISTS early_bird_discount_percentage INTEGER,
    ADD COLUMN IF NOT EXISTS group_discount_percentage INTEGER,
    ADD COLUMN IF NOT EXISTS group_min_attendees INTEGER;

-- 7. SEO & Marketing
ALTER TABLE events 
    ADD COLUMN IF NOT EXISTS seo_title VARCHAR(150),
    ADD COLUMN IF NOT EXISTS seo_description TEXT,
    ADD COLUMN IF NOT EXISTS seo_keywords TEXT[],
    ADD COLUMN IF NOT EXISTS seo_canonical_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS seo_robots VARCHAR(100),
    ADD COLUMN IF NOT EXISTS seo_noindex BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS og_title VARCHAR(150),
    ADD COLUMN IF NOT EXISTS og_description TEXT,
    ADD COLUMN IF NOT EXISTS og_image_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS og_type VARCHAR(50) DEFAULT 'event',
    ADD COLUMN IF NOT EXISTS twitter_card VARCHAR(50) DEFAULT 'summary_large_image',
    ADD COLUMN IF NOT EXISTS twitter_title VARCHAR(150),
    ADD COLUMN IF NOT EXISTS twitter_description TEXT,
    ADD COLUMN IF NOT EXISTS twitter_image_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS schema_org JSONB;

-- 8. Social & Engagement
ALTER TABLE events 
    ADD COLUMN IF NOT EXISTS social_links JSONB DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS has_livestream BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS livestream_url VARCHAR(500),
    ADD COLUMN IF NOT EXISTS recording_available BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS recording_url VARCHAR(500);

-- 9. Additional Metadata
ALTER TABLE events 
    ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1,
    ADD COLUMN IF NOT EXISTS published_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS scheduled_publish_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS last_published_at TIMESTAMP;

-- ============================================================
-- ADD INDEXES FOR PERFORMANCE
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_events_start_date ON events(start_date);
CREATE INDEX IF NOT EXISTS idx_events_end_date ON events(end_date);
CREATE INDEX IF NOT EXISTS idx_events_published_at ON events(published_at);
CREATE INDEX IF NOT EXISTS idx_events_scheduled_publish_at ON events(scheduled_publish_at);
CREATE INDEX IF NOT EXISTS idx_events_is_multi_day ON events(is_multi_day);
CREATE INDEX IF NOT EXISTS idx_events_is_hybrid ON events(is_hybrid);
CREATE INDEX IF NOT EXISTS idx_events_has_livestream ON events(has_livestream);
CREATE INDEX IF NOT EXISTS idx_events_recording_available ON events(recording_available);
CREATE INDEX IF NOT EXISTS idx_events_seo_noindex ON events(seo_noindex);
CREATE INDEX IF NOT EXISTS idx_events_ticket_sales_start ON events(ticket_sales_start);
CREATE INDEX IF NOT EXISTS idx_events_ticket_sales_end ON events(ticket_sales_end);
CREATE INDEX IF NOT EXISTS idx_events_language ON events(language);
CREATE INDEX IF NOT EXISTS idx_events_version ON events(version);
CREATE INDEX IF NOT EXISTS idx_events_featured_until ON events(featured_until);

-- ============================================================
-- CREATE SCHEDULES TABLE (for multi-day events)
-- ============================================================

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

-- ============================================================
-- CREATE TICKETS TABLE
-- ============================================================

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
CREATE INDEX IF NOT EXISTS idx_event_tickets_price ON event_tickets(price);

-- ============================================================
-- CREATE SPEAKERS TABLE
-- ============================================================

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
CREATE INDEX IF NOT EXISTS idx_event_speakers_is_keynote ON event_speakers(is_keynote);

-- ============================================================
-- CREATE MATERIALS TABLE
-- ============================================================

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

-- ============================================================
-- CREATE INVITATIONS TABLE (for invite-only events)
-- ============================================================

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
CREATE INDEX IF NOT EXISTS idx_event_invitations_accepted_at ON event_invitations(accepted_at);

-- ============================================================
-- CREATE EVENT REGISTRATIONS TABLE (for attendee tracking)
-- ============================================================

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

-- ============================================================
-- CREATE EVENT WAITLIST TABLE
-- ============================================================

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
-- COMMENTS FOR DOCUMENTATION
-- ============================================================

COMMENT ON TABLE event_schedules IS 'Schedules for multi-day events';
COMMENT ON TABLE event_tickets IS 'Ticket types and pricing for events';
COMMENT ON TABLE event_speakers IS 'Speakers and presenters for events';
COMMENT ON TABLE event_materials IS 'Materials and resources for events';
COMMENT ON TABLE event_invitations IS 'Invitations for invite-only events';
COMMENT ON TABLE event_registrations IS 'Event registrations and attendee tracking';
COMMENT ON TABLE event_waitlist IS 'Waitlist for sold-out events';

-- ============================================================
-- ADD FOREIGN KEY CONSTRAINTS (if not already added)
-- ============================================================

-- These might already exist from previous migrations, but ensure they're there
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_schedules_event_id' 
        AND table_name = 'event_schedules'
    ) THEN
        ALTER TABLE event_schedules 
            ADD CONSTRAINT fk_event_schedules_event_id 
            FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_tickets_event_id' 
        AND table_name = 'event_tickets'
    ) THEN
        ALTER TABLE event_tickets 
            ADD CONSTRAINT fk_event_tickets_event_id 
            FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_tickets_ticket_type_id' 
        AND table_name = 'event_tickets'
    ) THEN
        ALTER TABLE event_tickets 
            ADD CONSTRAINT fk_event_tickets_ticket_type_id 
            FOREIGN KEY (ticket_type_id) REFERENCES ticket_types(id) ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_speakers_event_id' 
        AND table_name = 'event_speakers'
    ) THEN
        ALTER TABLE event_speakers 
            ADD CONSTRAINT fk_event_speakers_event_id 
            FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_materials_event_id' 
        AND table_name = 'event_materials'
    ) THEN
        ALTER TABLE event_materials 
            ADD CONSTRAINT fk_event_materials_event_id 
            FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_materials_material_type_id' 
        AND table_name = 'event_materials'
    ) THEN
        ALTER TABLE event_materials 
            ADD CONSTRAINT fk_event_materials_material_type_id 
            FOREIGN KEY (material_type_id) REFERENCES material_types(id) ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_invitations_event_id' 
        AND table_name = 'event_invitations'
    ) THEN
        ALTER TABLE event_invitations 
            ADD CONSTRAINT fk_event_invitations_event_id 
            FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_invitations_invited_by' 
        AND table_name = 'event_invitations'
    ) THEN
        ALTER TABLE event_invitations 
            ADD CONSTRAINT fk_event_invitations_invited_by 
            FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_registrations_event_id' 
        AND table_name = 'event_registrations'
    ) THEN
        ALTER TABLE event_registrations 
            ADD CONSTRAINT fk_event_registrations_event_id 
            FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_registrations_user_id' 
        AND table_name = 'event_registrations'
    ) THEN
        ALTER TABLE event_registrations 
            ADD CONSTRAINT fk_event_registrations_user_id 
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_registrations_ticket_id' 
        AND table_name = 'event_registrations'
    ) THEN
        ALTER TABLE event_registrations 
            ADD CONSTRAINT fk_event_registrations_ticket_id 
            FOREIGN KEY (ticket_id) REFERENCES event_tickets(id) ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_waitlist_event_id' 
        AND table_name = 'event_waitlist'
    ) THEN
        ALTER TABLE event_waitlist 
            ADD CONSTRAINT fk_event_waitlist_event_id 
            FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_waitlist_user_id' 
        AND table_name = 'event_waitlist'
    ) THEN
        ALTER TABLE event_waitlist 
            ADD CONSTRAINT fk_event_waitlist_user_id 
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_event_waitlist_ticket_type_id' 
        AND table_name = 'event_waitlist'
    ) THEN
        ALTER TABLE event_waitlist 
            ADD CONSTRAINT fk_event_waitlist_ticket_type_id 
            FOREIGN KEY (ticket_type_id) REFERENCES ticket_types(id) ON DELETE SET NULL;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- ============================================================
-- DROP TABLES IN REVERSE ORDER (respecting foreign keys)
-- ============================================================

-- Drop waitlist table
DROP TABLE IF EXISTS event_waitlist;

-- Drop registrations table
DROP TABLE IF EXISTS event_registrations;

-- Drop invitations table
DROP TABLE IF EXISTS event_invitations;

-- Drop materials table
DROP TABLE IF EXISTS event_materials;

-- Drop speakers table
DROP TABLE IF EXISTS event_speakers;

-- Drop tickets table
DROP TABLE IF EXISTS event_tickets;

-- Drop schedules table
DROP TABLE IF EXISTS event_schedules;

-- ============================================================
-- REMOVE INDEXES
-- ============================================================

DROP INDEX IF EXISTS idx_events_start_date;
DROP INDEX IF EXISTS idx_events_end_date;
DROP INDEX IF EXISTS idx_events_published_at;
DROP INDEX IF EXISTS idx_events_scheduled_publish_at;
DROP INDEX IF EXISTS idx_events_is_multi_day;
DROP INDEX IF EXISTS idx_events_is_hybrid;
DROP INDEX IF EXISTS idx_events_has_livestream;
DROP INDEX IF EXISTS idx_events_recording_available;
DROP INDEX IF EXISTS idx_events_seo_noindex;
DROP INDEX IF EXISTS idx_events_ticket_sales_start;
DROP INDEX IF EXISTS idx_events_ticket_sales_end;
DROP INDEX IF EXISTS idx_events_language;
DROP INDEX IF EXISTS idx_events_version;
DROP INDEX IF EXISTS idx_events_featured_until;

DROP INDEX IF EXISTS idx_event_schedules_event_id;
DROP INDEX IF EXISTS idx_event_schedules_start_date;
DROP INDEX IF EXISTS idx_event_schedules_deleted_at;
DROP INDEX IF EXISTS idx_event_tickets_event_id;
DROP INDEX IF EXISTS idx_event_tickets_ticket_type_id;
DROP INDEX IF EXISTS idx_event_tickets_is_active;
DROP INDEX IF EXISTS idx_event_tickets_deleted_at;
DROP INDEX IF EXISTS idx_event_tickets_price;
DROP INDEX IF EXISTS idx_event_speakers_event_id;
DROP INDEX IF EXISTS idx_event_speakers_deleted_at;
DROP INDEX IF EXISTS idx_event_speakers_is_keynote;
DROP INDEX IF EXISTS idx_event_materials_event_id;
DROP INDEX IF EXISTS idx_event_materials_material_type_id;
DROP INDEX IF EXISTS idx_event_materials_is_pre_event;
DROP INDEX IF EXISTS idx_event_materials_deleted_at;
DROP INDEX IF EXISTS idx_event_invitations_event_id;
DROP INDEX IF EXISTS idx_event_invitations_email;
DROP INDEX IF EXISTS idx_event_invitations_token;
DROP INDEX IF EXISTS idx_event_invitations_deleted_at;
DROP INDEX IF EXISTS idx_event_invitations_accepted_at;
DROP INDEX IF EXISTS idx_event_registrations_event_id;
DROP INDEX IF EXISTS idx_event_registrations_user_id;
DROP INDEX IF EXISTS idx_event_registrations_ticket_id;
DROP INDEX IF EXISTS idx_event_registrations_status;
DROP INDEX IF EXISTS idx_event_registrations_registration_number;
DROP INDEX IF EXISTS idx_event_registrations_deleted_at;
DROP INDEX IF EXISTS idx_event_waitlist_event_id;
DROP INDEX IF EXISTS idx_event_waitlist_user_id;
DROP INDEX IF EXISTS idx_event_waitlist_position;
DROP INDEX IF EXISTS idx_event_waitlist_status;
DROP INDEX IF EXISTS idx_event_waitlist_deleted_at;

-- ============================================================
-- REMOVE COLUMNS FROM EVENTS TABLE
-- ============================================================

ALTER TABLE events 
    DROP COLUMN IF EXISTS tags,
    DROP COLUMN IF EXISTS short_description,
    DROP COLUMN IF EXISTS language,
    DROP COLUMN IF EXISTS start_date,
    DROP COLUMN IF EXISTS end_date,
    DROP COLUMN IF EXISTS is_multi_day,
    DROP COLUMN IF EXISTS recurrence_days_of_week,
    DROP COLUMN IF EXISTS recurrence_day_of_month,
    DROP COLUMN IF EXISTS recurrence_week_of_month,
    DROP COLUMN IF EXISTS venue_name,
    DROP COLUMN IF EXISTS venue_address,
    DROP COLUMN IF EXISTS venue_city,
    DROP COLUMN IF EXISTS venue_country,
    DROP COLUMN IF EXISTS venue_coordinates,
    DROP COLUMN IF EXISTS is_hybrid,
    DROP COLUMN IF EXISTS virtual_platform,
    DROP COLUMN IF EXISTS virtual_platform_url,
    DROP COLUMN IF EXISTS in_person_location,
    DROP COLUMN IF EXISTS waitlist_capacity,
    DROP COLUMN IF EXISTS ticket_sales_start,
    DROP COLUMN IF EXISTS ticket_sales_end,
    DROP COLUMN IF EXISTS min_tickets_per_order,
    DROP COLUMN IF EXISTS max_tickets_per_order,
    DROP COLUMN IF EXISTS invited_emails,
    DROP COLUMN IF EXISTS requires_approval,
    DROP COLUMN IF EXISTS approval_required_for,
    DROP COLUMN IF EXISTS featured_until,
    DROP COLUMN IF EXISTS early_bird_discount_percentage,
    DROP COLUMN IF EXISTS group_discount_percentage,
    DROP COLUMN IF EXISTS group_min_attendees,
    DROP COLUMN IF EXISTS seo_title,
    DROP COLUMN IF EXISTS seo_description,
    DROP COLUMN IF EXISTS seo_keywords,
    DROP COLUMN IF EXISTS seo_canonical_url,
    DROP COLUMN IF EXISTS seo_robots,
    DROP COLUMN IF EXISTS seo_noindex,
    DROP COLUMN IF EXISTS og_title,
    DROP COLUMN IF EXISTS og_description,
    DROP COLUMN IF EXISTS og_image_url,
    DROP COLUMN IF EXISTS og_type,
    DROP COLUMN IF EXISTS twitter_card,
    DROP COLUMN IF EXISTS twitter_title,
    DROP COLUMN IF EXISTS twitter_description,
    DROP COLUMN IF EXISTS twitter_image_url,
    DROP COLUMN IF EXISTS schema_org,
    DROP COLUMN IF EXISTS social_links,
    DROP COLUMN IF EXISTS has_livestream,
    DROP COLUMN IF EXISTS livestream_url,
    DROP COLUMN IF EXISTS recording_available,
    DROP COLUMN IF EXISTS recording_url,
    DROP COLUMN IF EXISTS metadata,
    DROP COLUMN IF EXISTS version,
    DROP COLUMN IF EXISTS published_at,
    DROP COLUMN IF EXISTS scheduled_publish_at,
    DROP COLUMN IF EXISTS last_published_at;
-- +goose StatementEnd