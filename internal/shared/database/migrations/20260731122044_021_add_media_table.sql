-- +goose Up
-- +goose StatementBegin

-- Create media table
CREATE TABLE IF NOT EXISTS media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    bucket VARCHAR(50) NOT NULL,
    path VARCHAR(500) NOT NULL,
    url VARCHAR(500) NOT NULL,
    media_type_id UUID NOT NULL,
    entity_id UUID NOT NULL,
    uploaded_by UUID NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_media_media_type_id ON media (media_type_id);
CREATE INDEX idx_media_entity ON media (entity_id);
CREATE INDEX idx_media_uploaded_by ON media (uploaded_by);
CREATE INDEX idx_media_is_active ON media (is_active);
CREATE INDEX idx_media_bucket ON media (bucket);

-- Add foreign key constraints
ALTER TABLE media ADD CONSTRAINT fk_media_media_type 
    FOREIGN KEY (media_type_id) REFERENCES media_types(id) ON DELETE RESTRICT;

ALTER TABLE media ADD CONSTRAINT fk_media_uploaded_by 
    FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE RESTRICT;

-- Create update trigger for updated_at
CREATE OR REPLACE FUNCTION update_media_updated_at() 
RETURNS TRIGGER LANGUAGE plpgsql AS '
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
';

DROP TRIGGER IF EXISTS update_media_updated_at ON media;
CREATE TRIGGER update_media_updated_at
    BEFORE UPDATE ON media
    FOR EACH ROW
    EXECUTE FUNCTION update_media_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop triggers
DROP TRIGGER IF EXISTS update_media_updated_at ON media;
DROP FUNCTION IF EXISTS update_media_updated_at();

-- Drop foreign key constraints
ALTER TABLE media DROP CONSTRAINT IF EXISTS fk_media_media_type;
ALTER TABLE media DROP CONSTRAINT IF EXISTS fk_media_uploaded_by;

-- Drop table
DROP TABLE IF EXISTS media;

-- +goose StatementEnd