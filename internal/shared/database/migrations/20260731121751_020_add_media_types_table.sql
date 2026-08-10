-- +goose Up
-- +goose StatementBegin

-- Create media_types table
CREATE TABLE IF NOT EXISTS media_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    bucket VARCHAR(50) NOT NULL,
    icon VARCHAR(50),
    max_file_size BIGINT DEFAULT 10485760,
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_media_types_name ON media_types (name);
CREATE INDEX idx_media_types_is_active ON media_types (is_active);
CREATE INDEX idx_media_types_sort_order ON media_types (sort_order);

-- Create update trigger for updated_at
CREATE OR REPLACE FUNCTION update_media_types_updated_at() 
RETURNS TRIGGER LANGUAGE plpgsql AS '
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
';

DROP TRIGGER IF EXISTS update_media_types_updated_at ON media_types;
CREATE TRIGGER update_media_types_updated_at
    BEFORE UPDATE ON media_types
    FOR EACH ROW
    EXECUTE FUNCTION update_media_types_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop trigger
DROP TRIGGER IF EXISTS update_media_types_updated_at ON media_types;
DROP FUNCTION IF EXISTS update_media_types_updated_at();

-- Drop table
DROP TABLE IF EXISTS media_types;

-- +goose StatementEnd