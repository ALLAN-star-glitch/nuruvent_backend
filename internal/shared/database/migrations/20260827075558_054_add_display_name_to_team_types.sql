-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- MIGRATION: Add display_name to team_types and reorder columns
-- ============================================================

-- ============================================================
-- 1. FIRST: ADD display_name column to existing table
-- ============================================================
ALTER TABLE team_types 
ADD COLUMN IF NOT EXISTS display_name VARCHAR(150);

-- ============================================================
-- 2. SECOND: UPDATE display_name with name value
-- ============================================================
UPDATE team_types 
SET display_name = name 
WHERE display_name IS NULL;

-- ============================================================
-- 3. THIRD: Make display_name NOT NULL
-- ============================================================
ALTER TABLE team_types 
ALTER COLUMN display_name SET NOT NULL;

-- ============================================================
-- 4. FOURTH: Recreate table with desired column order (optional)
-- ============================================================
-- Create new table with desired order
CREATE TABLE team_types_new (
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

-- Copy data from old table to new table
INSERT INTO team_types_new (id, name, display_name, slug, description, is_active, created_at, updated_at, deleted_at)
SELECT 
    id, 
    name, 
    display_name, 
    slug, 
    description, 
    is_active, 
    created_at, 
    updated_at, 
    deleted_at
FROM team_types;

-- Drop old table and rename new one
DROP TABLE team_types CASCADE;
ALTER TABLE team_types_new RENAME TO team_types;

-- Recreate indexes
CREATE INDEX idx_team_types_slug ON team_types(slug);
CREATE INDEX idx_team_types_is_active ON team_types(is_active);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- ============================================================
-- DOWN: Remove display_name and revert column order
-- ============================================================

-- Recreate table without display_name
CREATE TABLE team_types_old (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

INSERT INTO team_types_old (id, slug, name, description, is_active, created_at, updated_at, deleted_at)
SELECT id, slug, name, description, is_active, created_at, updated_at, deleted_at
FROM team_types;

DROP TABLE team_types CASCADE;
ALTER TABLE team_types_old RENAME TO team_types;

CREATE INDEX idx_team_types_slug ON team_types(slug);
CREATE INDEX idx_team_types_is_active ON team_types(is_active);

-- +goose StatementEnd