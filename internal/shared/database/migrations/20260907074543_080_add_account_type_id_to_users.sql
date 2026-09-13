-- +goose Up
-- +goose StatementBegin
-- Add account_type_id column to users table
ALTER TABLE users ADD COLUMN account_type_id UUID;

-- Add foreign key constraint to account_types table
ALTER TABLE users ADD CONSTRAINT fk_users_account_type_id 
    FOREIGN KEY (account_type_id) REFERENCES account_types(id) ON DELETE SET NULL;

-- Add index for better performance
CREATE INDEX idx_users_account_type_id ON users(account_type_id);

-- Update existing users to have a default account type
-- This ensures existing data is consistent
UPDATE users SET account_type_id = (
    SELECT id FROM account_types WHERE slug = 'personal' LIMIT 1
) WHERE account_type_id IS NULL;

-- Make the column NOT NULL after setting defaults
ALTER TABLE users ALTER COLUMN account_type_id SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove the column and its dependencies
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_account_type_id;
DROP INDEX IF EXISTS idx_users_account_type_id;
ALTER TABLE users DROP COLUMN account_type_id;
-- +goose StatementEnd