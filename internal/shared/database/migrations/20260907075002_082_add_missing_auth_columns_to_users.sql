-- +goose Up
-- +goose StatementBegin
-- Add identity_verified_at column (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'identity_verified_at'
    ) THEN
        ALTER TABLE users ADD COLUMN identity_verified_at TIMESTAMP WITH TIME ZONE;
        CREATE INDEX idx_users_identity_verified_at ON users(identity_verified_at);
    END IF;
END $$;

-- Add email_verified_at column (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'email_verified_at'
    ) THEN
        ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMP WITH TIME ZONE;
        CREATE INDEX idx_users_email_verified_at ON users(email_verified_at);
    END IF;
END $$;

-- Add phone_verified_at column (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'phone_verified_at'
    ) THEN
        ALTER TABLE users ADD COLUMN phone_verified_at TIMESTAMP WITH TIME ZONE;
        CREATE INDEX idx_users_phone_verified_at ON users(phone_verified_at);
    END IF;
END $$;

-- Add deleted_by column (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'deleted_by'
    ) THEN
        ALTER TABLE users ADD COLUMN deleted_by UUID;
        CREATE INDEX idx_users_deleted_by ON users(deleted_by);
    END IF;
END $$;

-- Add restored_at column (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'restored_at'
    ) THEN
        ALTER TABLE users ADD COLUMN restored_at TIMESTAMP WITH TIME ZONE;
        CREATE INDEX idx_users_restored_at ON users(restored_at);
    END IF;
END $$;

-- Add restored_by column (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'restored_by'
    ) THEN
        ALTER TABLE users ADD COLUMN restored_by UUID;
        CREATE INDEX idx_users_restored_by ON users(restored_by);
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_identity_verified_at;
DROP INDEX IF EXISTS idx_users_email_verified_at;
DROP INDEX IF EXISTS idx_users_phone_verified_at;
DROP INDEX IF EXISTS idx_users_deleted_by;
DROP INDEX IF EXISTS idx_users_restored_at;
DROP INDEX IF EXISTS idx_users_restored_by;

ALTER TABLE users DROP COLUMN IF EXISTS identity_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS phone_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_by;
ALTER TABLE users DROP COLUMN IF EXISTS restored_at;
ALTER TABLE users DROP COLUMN IF EXISTS restored_by;
-- +goose StatementEnd