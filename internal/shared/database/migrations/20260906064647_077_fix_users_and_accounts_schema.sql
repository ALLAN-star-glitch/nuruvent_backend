-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- STEP 1: Remove duplicate/incorrect columns from users
-- ============================================================

-- Drop indexes before dropping target columns
DROP INDEX IF EXISTS idx_users_account_type_id;
DROP INDEX IF EXISTS idx_users_identity_verified;
DROP INDEX IF EXISTS idx_users_kyc_status;

-- Clean up duplicate foreign keys on users
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_accounts_account_type;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_account_type;

-- Remove columns belonging to accounts
ALTER TABLE users 
    DROP COLUMN IF EXISTS account_type_id,
    DROP COLUMN IF EXISTS professional_type,
    DROP COLUMN IF EXISTS kyc_status,
    DROP COLUMN IF EXISTS kyc_submitted_at,
    DROP COLUMN IF EXISTS kyc_verified_at,
    DROP COLUMN IF EXISTS kyc_rejected_at,
    DROP COLUMN IF EXISTS kyc_rejection_reason,
    DROP COLUMN IF EXISTS id_document,
    DROP COLUMN IF EXISTS selfie_document,
    DROP COLUMN IF EXISTS address_proof,
    DROP COLUMN IF EXISTS identity_verified,
    DROP COLUMN IF EXISTS identity_verified_at;

-- ============================================================
-- STEP 2: Add display_name to accounts (optional, but useful)
-- ============================================================

ALTER TABLE accounts ADD COLUMN IF NOT EXISTS display_name VARCHAR(255);

-- ============================================================
-- STEP 3: Add KYC fields to accounts
-- ============================================================

ALTER TABLE accounts 
    ADD COLUMN IF NOT EXISTS kyc_status VARCHAR(50) DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS kyc_submitted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS kyc_verified_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS kyc_rejected_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS kyc_rejection_reason TEXT,
    ADD COLUMN IF NOT EXISTS business_registration VARCHAR(500),
    ADD COLUMN IF NOT EXISTS tax_id VARCHAR(500),
    ADD COLUMN IF NOT EXISTS directors_document VARCHAR(500),
    ADD COLUMN IF NOT EXISTS proof_of_address VARCHAR(500);

-- ============================================================
-- STEP 4: Add check constraint for kyc_status on accounts
-- ============================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'chk_accounts_kyc_status' 
        AND table_name = 'accounts'
    ) THEN
        ALTER TABLE accounts ADD CONSTRAINT chk_accounts_kyc_status 
            CHECK (kyc_status IN ('pending', 'submitted', 'verified', 'rejected', 'not_required'));
    END IF;
END $$;

-- ============================================================
-- STEP 5: Ensure account_type_id is set for all accounts
-- ============================================================

DO $$
DECLARE
    personal_type_id UUID;
BEGIN
    SELECT id INTO personal_type_id FROM account_types WHERE name = 'account_type_personal' LIMIT 1;
    
    IF personal_type_id IS NOT NULL THEN
        UPDATE accounts SET account_type_id = personal_type_id WHERE account_type_id IS NULL;
        RAISE NOTICE '✅ Set account_type_id for accounts with NULL';
    ELSE
        RAISE NOTICE '⚠️  account_types table does not have "account_type_personal". Please seed it first.';
    END IF;
END $$;

-- ============================================================
-- STEP 6: Add missing performance indexes
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_accounts_kyc_status ON accounts(kyc_status);
CREATE INDEX IF NOT EXISTS idx_accounts_type ON accounts(type);
CREATE INDEX IF NOT EXISTS idx_accounts_display_name ON accounts(display_name);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_professional_type_id ON users(professional_type_id);

-- ============================================================
-- STEP 7: Verification
-- ============================================================

DO $$
DECLARE
    col_count INTEGER;
    account_count INTEGER;
    null_type_count INTEGER;
BEGIN
    -- Check redundant columns removed from users
    SELECT COUNT(*) INTO col_count 
    FROM information_schema.columns 
    WHERE table_name = 'users' 
    AND column_name IN ('account_type_id', 'professional_type', 'kyc_status', 'identity_verified');
    
    IF col_count = 0 THEN
        RAISE NOTICE '✅ All redundant columns removed from users table.';
    ELSE
        RAISE NOTICE '⚠️ Some redundant columns still exist in users table.';
    END IF;
    
    -- Check accounts have account_type_id
    SELECT COUNT(*) INTO account_count FROM accounts;
    SELECT COUNT(*) INTO null_type_count FROM accounts WHERE account_type_id IS NULL;
    
    RAISE NOTICE '✅ Total accounts: %', account_count;
    RAISE NOTICE '✅ Accounts with account_type_id set: %', account_count - null_type_count;
    
    IF null_type_count > 0 THEN
        RAISE NOTICE '⚠️ % accounts still have NULL account_type_id', null_type_count;
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- ============================================================
-- ROLLBACK
-- ============================================================

-- 1. Drop added indexes first
DROP INDEX IF EXISTS idx_accounts_kyc_status;
DROP INDEX IF EXISTS idx_accounts_type;
DROP INDEX IF EXISTS idx_accounts_display_name;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_professional_type_id;

-- 2. Drop constraint on accounts before dropping columns
ALTER TABLE accounts DROP CONSTRAINT IF EXISTS chk_accounts_kyc_status;

-- 3. Remove KYC fields from accounts
ALTER TABLE accounts 
    DROP COLUMN IF EXISTS kyc_status,
    DROP COLUMN IF EXISTS kyc_submitted_at,
    DROP COLUMN IF EXISTS kyc_verified_at,
    DROP COLUMN IF EXISTS kyc_rejected_at,
    DROP COLUMN IF EXISTS kyc_rejection_reason,
    DROP COLUMN IF EXISTS business_registration,
    DROP COLUMN IF EXISTS tax_id,
    DROP COLUMN IF EXISTS directors_document,
    DROP COLUMN IF EXISTS proof_of_address,
    DROP COLUMN IF EXISTS display_name;

-- 4. Restore columns to users
ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS account_type_id UUID,
    ADD COLUMN IF NOT EXISTS professional_type VARCHAR(50),
    ADD COLUMN IF NOT EXISTS kyc_status VARCHAR(50) DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS kyc_submitted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS kyc_verified_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS kyc_rejected_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS kyc_rejection_reason TEXT,
    ADD COLUMN IF NOT EXISTS id_document VARCHAR(500),
    ADD COLUMN IF NOT EXISTS selfie_document VARCHAR(500),
    ADD COLUMN IF NOT EXISTS address_proof VARCHAR(500),
    ADD COLUMN IF NOT EXISTS identity_verified BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS identity_verified_at TIMESTAMPTZ;

-- 5. Restore index and FK constraint on users
CREATE INDEX IF NOT EXISTS idx_users_account_type_id ON users(account_type_id);

ALTER TABLE users ADD CONSTRAINT fk_users_account_type 
    FOREIGN KEY (account_type_id) REFERENCES account_types(id) ON DELETE SET NULL;

-- +goose StatementEnd