-- +goose Up
-- +goose StatementBegin
-- Add identity_verified column to users table
ALTER TABLE users ADD COLUMN identity_verified BOOLEAN DEFAULT false;

-- Add index for better performance
CREATE INDEX idx_users_identity_verified ON users(identity_verified);

-- Update existing users to have identity_verified = false
UPDATE users SET identity_verified = false WHERE identity_verified IS NULL;

-- Make the column NOT NULL after setting defaults
ALTER TABLE users ALTER COLUMN identity_verified SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove the column and its dependencies
DROP INDEX IF EXISTS idx_users_identity_verified;
ALTER TABLE users DROP COLUMN identity_verified;
-- +goose StatementEnd