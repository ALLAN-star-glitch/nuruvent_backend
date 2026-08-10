-- +goose Up
-- +goose StatementBegin

-- Make phone column nullable for organizations and users without phone
ALTER TABLE users ALTER COLUMN phone DROP NOT NULL;

-- Remove the existing unique constraint on phone
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_key;

-- Add a partial unique constraint that only applies to non-empty phone values
-- This allows multiple users with empty/null phone values
CREATE UNIQUE INDEX idx_users_phone_unique_not_empty ON users (phone) WHERE phone IS NOT NULL AND phone != '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop the partial unique index
DROP INDEX IF EXISTS idx_users_phone_unique_not_empty;

-- Re-add the original unique constraint
ALTER TABLE users ADD CONSTRAINT users_phone_key UNIQUE (phone);

-- Make phone column NOT NULL again
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;

-- +goose StatementEnd