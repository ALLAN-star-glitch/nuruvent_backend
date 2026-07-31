-- +goose Up
-- +goose StatementBegin

-- Add category column to business_types
ALTER TABLE business_types ADD COLUMN category VARCHAR(50) DEFAULT 'organization';

-- Add requires_business_name column
ALTER TABLE business_types ADD COLUMN requires_business_name BOOLEAN DEFAULT false;

-- Add requires_business_email column
ALTER TABLE business_types ADD COLUMN requires_business_email BOOLEAN DEFAULT false;

-- Update existing records
-- Organizations
UPDATE business_types SET 
    category = 'organization',
    requires_business_name = true,
    requires_business_email = true
WHERE name IN ('training_institute', 'college', 'professional_body', 'ngo');

-- Individual (existing)
UPDATE business_types SET 
    category = 'individual',
    requires_business_name = false,
    requires_business_email = false
WHERE name = 'individual';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove columns
ALTER TABLE business_types DROP COLUMN IF EXISTS category;
ALTER TABLE business_types DROP COLUMN IF EXISTS requires_business_name;
ALTER TABLE business_types DROP COLUMN IF EXISTS requires_business_email;

-- +goose StatementEnd