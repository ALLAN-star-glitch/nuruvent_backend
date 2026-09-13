-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- MIGRATION: Add missing KYC columns to users table
-- ============================================================

-- Only add columns that don't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'kyc_submitted_at') THEN
        ALTER TABLE users ADD COLUMN kyc_submitted_at TIMESTAMP;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'kyc_verified_at') THEN
        ALTER TABLE users ADD COLUMN kyc_verified_at TIMESTAMP;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'kyc_rejected_at') THEN
        ALTER TABLE users ADD COLUMN kyc_rejected_at TIMESTAMP;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'kyc_rejection_reason') THEN
        ALTER TABLE users ADD COLUMN kyc_rejection_reason TEXT;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'id_document') THEN
        ALTER TABLE users ADD COLUMN id_document VARCHAR(500);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'selfie_document') THEN
        ALTER TABLE users ADD COLUMN selfie_document VARCHAR(500);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'address_proof') THEN
        ALTER TABLE users ADD COLUMN address_proof VARCHAR(500);
    END IF;
END $$;

-- Add comments for the new columns
COMMENT ON COLUMN users.kyc_submitted_at IS 'When KYC documents were submitted';
COMMENT ON COLUMN users.kyc_verified_at IS 'When KYC was verified';
COMMENT ON COLUMN users.kyc_rejected_at IS 'When KYC was rejected';
COMMENT ON COLUMN users.kyc_rejection_reason IS 'Reason for KYC rejection';
COMMENT ON COLUMN users.id_document IS 'ID document URL (passport, national ID)';
COMMENT ON COLUMN users.selfie_document IS 'Selfie with ID document URL';
COMMENT ON COLUMN users.address_proof IS 'Proof of address document URL';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE users 
DROP COLUMN IF EXISTS kyc_submitted_at,
DROP COLUMN IF EXISTS kyc_verified_at,
DROP COLUMN IF EXISTS kyc_rejected_at,
DROP COLUMN IF EXISTS kyc_rejection_reason,
DROP COLUMN IF EXISTS id_document,
DROP COLUMN IF EXISTS selfie_document,
DROP COLUMN IF EXISTS address_proof;

-- +goose StatementEnd