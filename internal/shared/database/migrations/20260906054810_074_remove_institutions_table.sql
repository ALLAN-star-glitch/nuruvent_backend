-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- STEP 1: Remove institution_id from users (if exists)
-- ============================================================

ALTER TABLE users DROP COLUMN IF EXISTS institution_id;

-- ============================================================
-- STEP 2: Add account_type_id to accounts if it doesn't exist
-- ============================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounts' AND column_name = 'account_type_id'
    ) THEN
        ALTER TABLE accounts ADD COLUMN account_type_id UUID;
    END IF;
END $$;

-- ============================================================
-- STEP 3: Clean up redundant fields from accounts
-- ============================================================

ALTER TABLE accounts DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE accounts DROP COLUMN IF EXISTS display_name;

-- ============================================================
-- STEP 4: Add missing fields to accounts (for institutions)
-- ============================================================

ALTER TABLE accounts 
    ADD COLUMN IF NOT EXISTS address TEXT,
    ADD COLUMN IF NOT EXISTS city VARCHAR(100),
    ADD COLUMN IF NOT EXISTS country VARCHAR(100),
    ADD COLUMN IF NOT EXISTS website VARCHAR(255),
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS billing_email VARCHAR(255),
    ADD COLUMN IF NOT EXISTS subscription_plan VARCHAR(50);

-- ============================================================
-- STEP 5: Set account_type_id for all accounts
-- ============================================================

-- First, try to get personal account type ID
DO $$
DECLARE
    personal_type_id UUID;
BEGIN
    SELECT id INTO personal_type_id FROM account_types WHERE name = 'account_type_personal' LIMIT 1;
    
    IF personal_type_id IS NOT NULL THEN
        UPDATE accounts SET account_type_id = personal_type_id WHERE account_type_id IS NULL;
        RAISE NOTICE '✅ Set account_type_id to personal for accounts with NULL';
    ELSE
        RAISE NOTICE '⚠️  account_types table does not have "account_type_personal". Please seed it first.';
    END IF;
END $$;

-- ============================================================
-- STEP 6: Drop institutions table
-- ============================================================

DROP TABLE IF EXISTS institutions CASCADE;

-- ============================================================
-- STEP 7: Add indexes for performance
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_accounts_account_type_id ON accounts(account_type_id);
CREATE INDEX IF NOT EXISTS idx_accounts_slug ON accounts(slug);
CREATE INDEX IF NOT EXISTS idx_accounts_status ON accounts(status);
CREATE INDEX IF NOT EXISTS idx_accounts_deleted_at ON accounts(deleted_at);

-- ============================================================
-- STEP 8: Verify migration
-- ============================================================

DO $$
DECLARE
    invalid_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO invalid_count FROM accounts WHERE account_type_id IS NULL;
    IF invalid_count > 0 THEN
        RAISE NOTICE '⚠️  Found % accounts with NULL account_type_id. Please fix manually.', invalid_count;
    ELSE
        RAISE NOTICE '✅ All accounts have valid account_type_id';
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- ============================================================
-- ROLLBACK
-- ============================================================

-- Restore institutions table
CREATE TABLE IF NOT EXISTS institutions (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    slug VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(50),
    logo_url VARCHAR(500),
    status VARCHAR(50) DEFAULT 'active',
    website VARCHAR(255),
    description TEXT,
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100),
    institution_type VARCHAR(50),
    kyc_status VARCHAR(50) DEFAULT 'pending',
    verified_at TIMESTAMP,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Restore institution_id to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS institution_id UUID;

-- Restore avatar_url to accounts
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);

-- Restore display_name to accounts
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS display_name VARCHAR(255);

-- Remove added fields
ALTER TABLE accounts 
    DROP COLUMN IF EXISTS address,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS country,
    DROP COLUMN IF EXISTS website,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS billing_email,
    DROP COLUMN IF EXISTS subscription_plan;

-- Remove account_type_id
ALTER TABLE accounts DROP COLUMN IF EXISTS account_type_id;

-- Drop indexes
DROP INDEX IF EXISTS idx_accounts_account_type_id;
DROP INDEX IF EXISTS idx_accounts_slug;
DROP INDEX IF EXISTS idx_accounts_status;
DROP INDEX IF EXISTS idx_accounts_deleted_at;

-- +goose StatementEnd