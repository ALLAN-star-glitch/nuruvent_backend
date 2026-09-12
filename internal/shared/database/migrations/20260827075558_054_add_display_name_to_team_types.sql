-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- MIGRATION: Create team_types (with display_name)
-- ============================================================
--
-- NOTE: This migration was originally written to ALTER an existing
-- team_types table. Migration 052, which was supposed to create it,
-- was corrupted and replaced with a no-op. This migration therefore
-- creates the table directly with the target schema.
--
-- The schema matches what the original rebuild sequence was trying to
-- produce (columns in the desired order, display_name present).

CREATE TABLE IF NOT EXISTS team_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(150) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX IF NOT EXISTS idx_team_types_slug ON team_types(slug);
CREATE INDEX IF NOT EXISTS idx_team_types_is_active ON team_types(is_active);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS team_types;
-- +goose StatementEnd