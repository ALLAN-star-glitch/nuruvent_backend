-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- MIGRATION: Add KYC columns to institutions table
-- ============================================================

ALTER TABLE institutions 
ADD COLUMN kyc_status VARCHAR(50) DEFAULT 'pending',
ADD COLUMN kyc_submitted_at TIMESTAMP,
ADD COLUMN kyc_verified_at TIMESTAMP,
ADD COLUMN kyc_rejected_at TIMESTAMP,
ADD COLUMN kyc_rejection_reason TEXT,
ADD COLUMN business_registration_number VARCHAR(100),
ADD COLUMN tax_pin VARCHAR(100),
ADD COLUMN business_license VARCHAR(500),
ADD COLUMN cr12_document VARCHAR(500),
ADD COLUMN directors_document VARCHAR(500),
ADD COLUMN verified_by UUID REFERENCES users(id);

CREATE INDEX idx_institutions_kyc_status ON institutions(kyc_status);

ALTER TABLE institutions 
ADD CONSTRAINT check_institution_kyc_status 
CHECK (kyc_status IN ('pending', 'submitted', 'verified', 'rejected', 'not_required'));

COMMENT ON COLUMN institutions.kyc_status IS 'KYC verification status';
COMMENT ON COLUMN institutions.kyc_submitted_at IS 'When KYC documents were submitted';
COMMENT ON COLUMN institutions.kyc_verified_at IS 'When KYC was verified';
COMMENT ON COLUMN institutions.kyc_rejected_at IS 'When KYC was rejected';
COMMENT ON COLUMN institutions.kyc_rejection_reason IS 'Reason for KYC rejection';
COMMENT ON COLUMN institutions.business_registration_number IS 'Business registration number';
COMMENT ON COLUMN institutions.tax_pin IS 'KRA Tax PIN';
COMMENT ON COLUMN institutions.business_license IS 'Business license document URL';
COMMENT ON COLUMN institutions.cr12_document IS 'Certificate of Incorporation (CR12) document URL';
COMMENT ON COLUMN institutions.directors_document IS 'Directors details document URL';
COMMENT ON COLUMN institutions.verified_by IS 'Admin user ID who verified KYC';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE institutions 
DROP COLUMN IF EXISTS kyc_status,
DROP COLUMN IF EXISTS kyc_submitted_at,
DROP COLUMN IF EXISTS kyc_verified_at,
DROP COLUMN IF EXISTS kyc_rejected_at,
DROP COLUMN IF EXISTS kyc_rejection_reason,
DROP COLUMN IF EXISTS business_registration_number,
DROP COLUMN IF EXISTS tax_pin,
DROP COLUMN IF EXISTS business_license,
DROP COLUMN IF EXISTS cr12_document,
DROP COLUMN IF EXISTS directors_document,
DROP COLUMN IF EXISTS verified_by;

DROP INDEX IF EXISTS idx_institutions_kyc_status;

-- +goose StatementEnd