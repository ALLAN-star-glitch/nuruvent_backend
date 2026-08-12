-- +goose Up
-- +goose StatementBegin

-- 1. Rename password to password_hash
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounts' AND column_name = 'password'
    ) THEN
        ALTER TABLE accounts RENAME COLUMN password TO password_hash;
    END IF;
END $$;

-- 2. Add phone_verified column
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounts' AND column_name = 'phone_verified'
    ) THEN
        ALTER TABLE accounts ADD COLUMN phone_verified BOOLEAN DEFAULT FALSE;
    END IF;
END $$;

-- 3. Add phone_verified_at column
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounts' AND column_name = 'phone_verified_at'
    ) THEN
        ALTER TABLE accounts ADD COLUMN phone_verified_at TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;

-- 4. Add kyc_status column
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounts' AND column_name = 'kyc_status'
    ) THEN
        ALTER TABLE accounts ADD COLUMN kyc_status VARCHAR(50) DEFAULT 'pending';
    END IF;
END $$;

-- 5. Add kyc_status constraint
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_kyc_status'
    ) THEN
        ALTER TABLE accounts ADD CONSTRAINT chk_kyc_status 
            CHECK (kyc_status IN ('pending', 'submitted', 'verified', 'rejected', 'not_required'));
    END IF;
END $$;

-- 6. Add indexes
CREATE INDEX IF NOT EXISTS idx_accounts_phone_verified ON accounts(phone_verified);
CREATE INDEX IF NOT EXISTS idx_accounts_kyc_status ON accounts(kyc_status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_accounts_phone_verified;
DROP INDEX IF EXISTS idx_accounts_kyc_status;

ALTER TABLE accounts DROP CONSTRAINT IF EXISTS chk_kyc_status;

ALTER TABLE accounts DROP COLUMN IF EXISTS kyc_status;
ALTER TABLE accounts DROP COLUMN IF EXISTS phone_verified_at;
ALTER TABLE accounts DROP COLUMN IF EXISTS phone_verified;

ALTER TABLE accounts RENAME COLUMN password_hash TO password;

-- +goose StatementEnd