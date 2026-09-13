-- +goose Up
-- +goose StatementBegin
-- Add is_active column to accounts table
ALTER TABLE accounts ADD COLUMN is_active BOOLEAN DEFAULT true;

-- Update existing records to have is_active = true
UPDATE accounts SET is_active = true WHERE is_active IS NULL;

-- Add index on is_active for better query performance
CREATE INDEX idx_accounts_is_active ON accounts(is_active);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove is_active column from accounts table
DROP INDEX IF EXISTS idx_accounts_is_active;
ALTER TABLE accounts DROP COLUMN is_active;
-- +goose StatementEnd