-- +goose Up
-- Remove can_host column from professional_types table
ALTER TABLE professional_types DROP COLUMN IF EXISTS can_host;

-- +goose Down
-- Add back can_host column with default value
ALTER TABLE professional_types ADD COLUMN IF NOT EXISTS can_host BOOLEAN DEFAULT FALSE;

-- Restore previous values based on slug (optional - uncomment if needed)
-- UPDATE professional_types SET can_host = TRUE 
-- WHERE slug IN ('professional-type-trainer', 'professional-type-coach', 'professional-type-consultant');
-- UPDATE professional_types SET can_host = FALSE 
-- WHERE slug = 'professional-type-freelancer';