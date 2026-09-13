-- +goose Up
-- Add avatar_url column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);

-- Add index for faster queries
CREATE INDEX IF NOT EXISTS idx_users_avatar_url ON users(avatar_url);

-- +goose Down
-- Remove avatar_url column from users table
DROP INDEX IF EXISTS idx_users_avatar_url;
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;