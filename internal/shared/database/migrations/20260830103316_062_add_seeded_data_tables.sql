-- +goose Up
-- +goose StatementBegin
-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- +goose StatementEnd

-- ============================================================
-- EVENT FORMATS TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS event_formats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    icon VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_formats_slug ON event_formats(slug);
CREATE INDEX IF NOT EXISTS idx_event_formats_is_active ON event_formats(is_active);
CREATE INDEX IF NOT EXISTS idx_event_formats_deleted_at ON event_formats(deleted_at);

-- ============================================================
-- CATEGORIES TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(20),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug);
CREATE INDEX IF NOT EXISTS idx_categories_sort_order ON categories(sort_order);
CREATE INDEX IF NOT EXISTS idx_categories_is_active ON categories(is_active);
CREATE INDEX IF NOT EXISTS idx_categories_deleted_at ON categories(deleted_at);

-- ============================================================
-- RECURRENCE PATTERNS TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS recurrence_patterns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_recurrence_patterns_slug ON recurrence_patterns(slug);
CREATE INDEX IF NOT EXISTS idx_recurrence_patterns_is_active ON recurrence_patterns(is_active);
CREATE INDEX IF NOT EXISTS idx_recurrence_patterns_deleted_at ON recurrence_patterns(deleted_at);

-- ============================================================
-- CERTIFICATE TEMPLATES TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS certificate_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    preview_url VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_certificate_templates_slug ON certificate_templates(slug);
CREATE INDEX IF NOT EXISTS idx_certificate_templates_is_active ON certificate_templates(is_active);
CREATE INDEX IF NOT EXISTS idx_certificate_templates_deleted_at ON certificate_templates(deleted_at);

-- ============================================================
-- CERTIFICATE TYPES TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS certificate_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_certificate_types_slug ON certificate_types(slug);
CREATE INDEX IF NOT EXISTS idx_certificate_types_is_active ON certificate_types(is_active);
CREATE INDEX IF NOT EXISTS idx_certificate_types_deleted_at ON certificate_types(deleted_at);

-- ============================================================
-- MATERIAL TYPES TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS material_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    icon VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_material_types_slug ON material_types(slug);
CREATE INDEX IF NOT EXISTS idx_material_types_is_active ON material_types(is_active);
CREATE INDEX IF NOT EXISTS idx_material_types_deleted_at ON material_types(deleted_at);

-- ============================================================
-- TICKET TYPES TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS ticket_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ticket_types_slug ON ticket_types(slug);
CREATE INDEX IF NOT EXISTS idx_ticket_types_sort_order ON ticket_types(sort_order);
CREATE INDEX IF NOT EXISTS idx_ticket_types_is_active ON ticket_types(is_active);
CREATE INDEX IF NOT EXISTS idx_ticket_types_deleted_at ON ticket_types(deleted_at);

-- ============================================================
-- EVENT CATEGORY JUNCTION TABLE (Many-to-Many)
-- ============================================================
-- This allows events to have multiple categories
CREATE TABLE IF NOT EXISTS event_categories (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id, category_id)
);

CREATE INDEX IF NOT EXISTS idx_event_categories_event_id ON event_categories(event_id);
CREATE INDEX IF NOT EXISTS idx_event_categories_category_id ON event_categories(category_id);

-- ============================================================
-- EVENT FORMAT JUNCTION TABLE
-- ============================================================
-- This links events to their formats
CREATE TABLE IF NOT EXISTS event_format_links (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    format_id UUID NOT NULL REFERENCES event_formats(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id, format_id)
);

CREATE INDEX IF NOT EXISTS idx_event_format_links_event_id ON event_format_links(event_id);
CREATE INDEX IF NOT EXISTS idx_event_format_links_format_id ON event_format_links(format_id);

-- ============================================================
-- ADD COLUMNS TO EVENTS TABLE
-- ============================================================
ALTER TABLE events 
    ADD COLUMN IF NOT EXISTS category_id UUID,
    ADD COLUMN IF NOT EXISTS event_format_id UUID,
    ADD COLUMN IF NOT EXISTS certificate_template_id UUID,
    ADD COLUMN IF NOT EXISTS is_virtual BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS is_recurring BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS recurrence_pattern_id UUID,
    ADD COLUMN IF NOT EXISTS recurrence_interval INTEGER DEFAULT 1,
    ADD COLUMN IF NOT EXISTS recurrence_ends_on TIMESTAMP,
    ADD COLUMN IF NOT EXISTS recurrence_occurrences INTEGER,
    ADD COLUMN IF NOT EXISTS timezone VARCHAR(50) DEFAULT 'Africa/Nairobi',
    ADD COLUMN IF NOT EXISTS waitlist_enabled BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS invite_only BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS certificate_enabled BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS is_free BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS capacity INTEGER,
    ADD COLUMN IF NOT EXISTS visibility VARCHAR(20) DEFAULT 'public',
    ADD COLUMN IF NOT EXISTS password VARCHAR(255);

-- ============================================================
-- ADD FOREIGN KEY CONSTRAINTS
-- ============================================================
-- Add foreign key constraints (without DO $$ block)
ALTER TABLE events ADD CONSTRAINT fk_events_category_id 
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;

ALTER TABLE events ADD CONSTRAINT fk_events_event_format_id 
    FOREIGN KEY (event_format_id) REFERENCES event_formats(id) ON DELETE SET NULL;

ALTER TABLE events ADD CONSTRAINT fk_events_certificate_template_id 
    FOREIGN KEY (certificate_template_id) REFERENCES certificate_templates(id) ON DELETE SET NULL;

ALTER TABLE events ADD CONSTRAINT fk_events_recurrence_pattern_id 
    FOREIGN KEY (recurrence_pattern_id) REFERENCES recurrence_patterns(id) ON DELETE SET NULL;

-- ============================================================
-- ADD INDEXES FOR PERFORMANCE
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_events_category_id ON events(category_id);
CREATE INDEX IF NOT EXISTS idx_events_event_format_id ON events(event_format_id);
CREATE INDEX IF NOT EXISTS idx_events_certificate_template_id ON events(certificate_template_id);
CREATE INDEX IF NOT EXISTS idx_events_is_virtual ON events(is_virtual);
CREATE INDEX IF NOT EXISTS idx_events_is_recurring ON events(is_recurring);
CREATE INDEX IF NOT EXISTS idx_events_visibility ON events(visibility);
CREATE INDEX IF NOT EXISTS idx_events_is_free ON events(is_free);
CREATE INDEX IF NOT EXISTS idx_events_certificate_enabled ON events(certificate_enabled);

-- ============================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================
COMMENT ON TABLE event_formats IS 'Event formats: Virtual, In-Person, Hybrid';
COMMENT ON TABLE categories IS 'Event categories for organization and discovery';
COMMENT ON TABLE recurrence_patterns IS 'Recurrence patterns: Daily, Weekly, Monthly, Custom';
COMMENT ON TABLE certificate_templates IS 'Certificate templates for event certificates';
COMMENT ON TABLE certificate_types IS 'Certificate types: Event, Course, CPD';
COMMENT ON TABLE material_types IS 'Material types: PDF, Video, Link, Document';
COMMENT ON TABLE ticket_types IS 'Ticket types: General, Early Bird, VIP, Group';

-- +goose Down
-- +goose StatementBegin
-- Drop foreign key constraints first
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_category_id;
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_event_format_id;
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_certificate_template_id;
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_recurrence_pattern_id;

ALTER TABLE event_format_links DROP CONSTRAINT IF EXISTS event_format_links_event_id_fkey;
ALTER TABLE event_format_links DROP CONSTRAINT IF EXISTS event_format_links_format_id_fkey;
ALTER TABLE event_categories DROP CONSTRAINT IF EXISTS event_categories_event_id_fkey;
ALTER TABLE event_categories DROP CONSTRAINT IF EXISTS event_categories_category_id_fkey;

-- Drop junction tables
DROP TABLE IF EXISTS event_format_links;
DROP TABLE IF EXISTS event_categories;

-- Drop main tables
DROP TABLE IF EXISTS ticket_types;
DROP TABLE IF EXISTS material_types;
DROP TABLE IF EXISTS certificate_types;
DROP TABLE IF EXISTS certificate_templates;
DROP TABLE IF EXISTS recurrence_patterns;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS event_formats;

-- Remove columns added to events table
ALTER TABLE events 
    DROP COLUMN IF EXISTS category_id,
    DROP COLUMN IF EXISTS event_format_id,
    DROP COLUMN IF EXISTS certificate_template_id,
    DROP COLUMN IF EXISTS is_virtual,
    DROP COLUMN IF EXISTS is_recurring,
    DROP COLUMN IF EXISTS recurrence_pattern_id,
    DROP COLUMN IF EXISTS recurrence_interval,
    DROP COLUMN IF EXISTS recurrence_ends_on,
    DROP COLUMN IF EXISTS recurrence_occurrences,
    DROP COLUMN IF EXISTS timezone,
    DROP COLUMN IF EXISTS waitlist_enabled,
    DROP COLUMN IF EXISTS invite_only,
    DROP COLUMN IF EXISTS certificate_enabled,
    DROP COLUMN IF EXISTS is_free,
    DROP COLUMN IF EXISTS capacity,
    DROP COLUMN IF EXISTS visibility,
    DROP COLUMN IF EXISTS password;

-- Drop indexes
DROP INDEX IF EXISTS idx_event_formats_slug;
DROP INDEX IF EXISTS idx_event_formats_is_active;
DROP INDEX IF EXISTS idx_event_formats_deleted_at;
DROP INDEX IF EXISTS idx_categories_slug;
DROP INDEX IF EXISTS idx_categories_sort_order;
DROP INDEX IF EXISTS idx_categories_is_active;
DROP INDEX IF EXISTS idx_categories_deleted_at;
DROP INDEX IF EXISTS idx_recurrence_patterns_slug;
DROP INDEX IF EXISTS idx_recurrence_patterns_is_active;
DROP INDEX IF EXISTS idx_recurrence_patterns_deleted_at;
DROP INDEX IF EXISTS idx_certificate_templates_slug;
DROP INDEX IF EXISTS idx_certificate_templates_is_active;
DROP INDEX IF EXISTS idx_certificate_templates_deleted_at;
DROP INDEX IF EXISTS idx_certificate_types_slug;
DROP INDEX IF EXISTS idx_certificate_types_is_active;
DROP INDEX IF EXISTS idx_certificate_types_deleted_at;
DROP INDEX IF EXISTS idx_material_types_slug;
DROP INDEX IF EXISTS idx_material_types_is_active;
DROP INDEX IF EXISTS idx_material_types_deleted_at;
DROP INDEX IF EXISTS idx_ticket_types_slug;
DROP INDEX IF EXISTS idx_ticket_types_sort_order;
DROP INDEX IF EXISTS idx_ticket_types_is_active;
DROP INDEX IF EXISTS idx_ticket_types_deleted_at;
DROP INDEX IF EXISTS idx_event_categories_event_id;
DROP INDEX IF EXISTS idx_event_categories_category_id;
DROP INDEX IF EXISTS idx_event_format_links_event_id;
DROP INDEX IF EXISTS idx_event_format_links_format_id;
DROP INDEX IF EXISTS idx_events_category_id;
DROP INDEX IF EXISTS idx_events_event_format_id;
DROP INDEX IF EXISTS idx_events_certificate_template_id;
DROP INDEX IF EXISTS idx_events_is_virtual;
DROP INDEX IF EXISTS idx_events_is_recurring;
DROP INDEX IF EXISTS idx_events_visibility;
DROP INDEX IF EXISTS idx_events_is_free;
DROP INDEX IF EXISTS idx_events_certificate_enabled;
-- +goose StatementEnd