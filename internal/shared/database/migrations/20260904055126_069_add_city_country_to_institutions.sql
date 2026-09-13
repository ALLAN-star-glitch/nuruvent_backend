-- +goose Up
-- Add city and country columns to institutions table
ALTER TABLE institutions ADD COLUMN IF NOT EXISTS city VARCHAR(255);
ALTER TABLE institutions ADD COLUMN IF NOT EXISTS country VARCHAR(255);

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_institutions_city ON institutions(city);
CREATE INDEX IF NOT EXISTS idx_institutions_country ON institutions(country);

-- +goose Down
-- Remove city and country columns from institutions table
DROP INDEX IF EXISTS idx_institutions_country;
DROP INDEX IF EXISTS idx_institutions_city;
ALTER TABLE institutions DROP COLUMN IF EXISTS country;
ALTER TABLE institutions DROP COLUMN IF EXISTS city;