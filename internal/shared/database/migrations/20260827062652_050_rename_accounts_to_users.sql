-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- MIGRATION: Rename accounts table to users
-- Date: 2026-08-27
-- Description: Rename accounts → users for clarity
-- ============================================================

-- ============================================================
-- 1. RENAME accounts → users
-- ============================================================
ALTER TABLE accounts RENAME TO users;

-- ============================================================
-- 2. UPDATE FOREIGN KEY REFERENCES
-- ============================================================
-- Rename account_id to user_id in team_members
ALTER TABLE team_members RENAME COLUMN account_id TO user_id;

-- Rename account_id in other tables (if they exist)
-- ALTER TABLE events RENAME COLUMN account_id TO user_id;
-- ALTER TABLE refresh_tokens RENAME COLUMN account_id TO user_id;

-- ============================================================
-- 3. RENAME FOREIGN KEY CONSTRAINTS (optional)
-- ============================================================
-- Note: PostgreSQL automatically updates constraint names
-- But you may want to rename them for clarity

ALTER TABLE team_members RENAME CONSTRAINT fk_team_members_account TO fk_team_members_user;
ALTER TABLE team_members RENAME CONSTRAINT fk_team_members_creator TO fk_team_members_creator_user;
ALTER TABLE team_members RENAME CONSTRAINT fk_team_members_member TO fk_team_members_member_user;

-- If events table exists:
-- ALTER TABLE events RENAME CONSTRAINT fk_events_account TO fk_events_user;

-- If refresh_tokens table exists:
-- ALTER TABLE refresh_tokens RENAME CONSTRAINT fk_refresh_tokens_account TO fk_refresh_tokens_user;

-- ============================================================
-- 4. RENAME INDEXES (optional)
-- ============================================================
ALTER INDEX accounts_pkey RENAME TO users_pkey;
ALTER INDEX accounts_email_key RENAME TO users_email_key;
ALTER INDEX accounts_slug_key RENAME TO users_slug_key;
ALTER INDEX idx_accounts_account_type_id RENAME TO idx_users_account_type_id;
ALTER INDEX idx_accounts_email RENAME TO idx_users_email;
ALTER INDEX idx_accounts_email_verified RENAME TO idx_users_email_verified;
ALTER INDEX idx_accounts_identity_verified RENAME TO idx_users_identity_verified;
ALTER INDEX idx_accounts_institution_id RENAME TO idx_users_institution_id;
ALTER INDEX idx_accounts_is_active RENAME TO idx_users_is_active;
ALTER INDEX idx_accounts_kyc_status RENAME TO idx_users_kyc_status;
ALTER INDEX idx_accounts_phone_verified RENAME TO idx_users_phone_verified;
ALTER INDEX idx_accounts_professional_type_id RENAME TO idx_users_professional_type_id;
ALTER INDEX idx_accounts_slug RENAME TO idx_users_slug;

-- ============================================================
-- 5. RENAME TRIGGER
-- ============================================================
ALTER TRIGGER update_accounts_updated_at ON users RENAME TO update_users_updated_at;

-- Note: The trigger function name can stay as is, or rename it too:
-- ALTER FUNCTION update_accounts_updated_at RENAME TO update_users_updated_at;

-- ============================================================
-- 6. UPDATE COMMENTS
-- ============================================================
COMMENT ON TABLE users IS 'Registered users of the platform';
COMMENT ON COLUMN users.id IS 'Unique user identifier';
COMMENT ON COLUMN users.email IS 'User email address (login credential)';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password';
COMMENT ON COLUMN users.name IS 'User full name';
COMMENT ON COLUMN users.phone IS 'User phone number (M-Pesa contact)';
COMMENT ON COLUMN users.professional_type IS 'Professional type (trainer, coach, consultant, freelancer)';
COMMENT ON COLUMN users.institution_id IS 'Institution this user belongs to (NULL for independent users)';
COMMENT ON COLUMN users.email_verified IS 'Whether the email has been verified';
COMMENT ON COLUMN users.email_verified_at IS 'Timestamp when email was verified';
COMMENT ON COLUMN users.identity_verified IS 'Whether identity has been verified';
COMMENT ON COLUMN users.identity_verified_at IS 'Timestamp when identity was verified';
COMMENT ON COLUMN users.is_active IS 'Whether the user account is active';
COMMENT ON COLUMN users.created_at IS 'Account creation timestamp';
COMMENT ON COLUMN users.updated_at IS 'Last update timestamp';
COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp (NULL if not deleted)';
COMMENT ON COLUMN users.account_type_id IS 'Type of account (personal or institution)';
COMMENT ON COLUMN users.professional_type_id IS 'Professional type reference (trainer, coach, etc.)';
COMMENT ON COLUMN users.slug IS 'URL-friendly unique identifier';
COMMENT ON COLUMN users.display_name IS 'User display name for UI';
COMMENT ON COLUMN users.phone_verified IS 'Whether phone has been verified';
COMMENT ON COLUMN users.phone_verified_at IS 'Timestamp when phone was verified';
COMMENT ON COLUMN users.kyc_status IS 'KYC verification status: pending, submitted, verified, rejected, not_required';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- ============================================================
-- DOWN: Revert rename accounts → users
-- ============================================================

-- Rename users back to accounts
ALTER TABLE users RENAME TO accounts;

-- Rename user_id back to account_id in team_members
ALTER TABLE team_members RENAME COLUMN user_id TO account_id;

-- Revert foreign key constraint names
ALTER TABLE team_members RENAME CONSTRAINT fk_team_members_user TO fk_team_members_account;
ALTER TABLE team_members RENAME CONSTRAINT fk_team_members_creator_user TO fk_team_members_creator;
ALTER TABLE team_members RENAME CONSTRAINT fk_team_members_member_user TO fk_team_members_member;

-- Revert indexes
ALTER INDEX users_pkey RENAME TO accounts_pkey;
ALTER INDEX users_email_key RENAME TO accounts_email_key;
ALTER INDEX users_slug_key RENAME TO accounts_slug_key;
ALTER INDEX idx_users_account_type_id RENAME TO idx_accounts_account_type_id;
ALTER INDEX idx_users_email RENAME TO idx_accounts_email;
ALTER INDEX idx_users_email_verified RENAME TO idx_accounts_email_verified;
ALTER INDEX idx_users_identity_verified RENAME TO idx_accounts_identity_verified;
ALTER INDEX idx_users_institution_id RENAME TO idx_accounts_institution_id;
ALTER INDEX idx_users_is_active RENAME TO idx_accounts_is_active;
ALTER INDEX idx_users_kyc_status RENAME TO idx_accounts_kyc_status;
ALTER INDEX idx_users_phone_verified RENAME TO idx_accounts_phone_verified;
ALTER INDEX idx_users_professional_type_id RENAME TO idx_accounts_professional_type_id;
ALTER INDEX idx_users_slug RENAME TO idx_accounts_slug;

-- Revert trigger
ALTER TRIGGER update_users_updated_at ON accounts RENAME TO update_accounts_updated_at;

-- Revert comments
COMMENT ON TABLE accounts IS NULL;
COMMENT ON COLUMN accounts.id IS NULL;
COMMENT ON COLUMN accounts.email IS NULL;
COMMENT ON COLUMN accounts.password_hash IS NULL;
COMMENT ON COLUMN accounts.name IS NULL;
COMMENT ON COLUMN accounts.phone IS NULL;
COMMENT ON COLUMN accounts.professional_type IS NULL;
COMMENT ON COLUMN accounts.institution_id IS NULL;
COMMENT ON COLUMN accounts.email_verified IS NULL;
COMMENT ON COLUMN accounts.email_verified_at IS NULL;
COMMENT ON COLUMN accounts.identity_verified IS NULL;
COMMENT ON COLUMN accounts.identity_verified_at IS NULL;
COMMENT ON COLUMN accounts.is_active IS NULL;
COMMENT ON COLUMN accounts.created_at IS NULL;
COMMENT ON COLUMN accounts.updated_at IS NULL;
COMMENT ON COLUMN accounts.deleted_at IS NULL;
COMMENT ON COLUMN accounts.account_type_id IS NULL;
COMMENT ON COLUMN accounts.professional_type_id IS NULL;
COMMENT ON COLUMN accounts.slug IS NULL;
COMMENT ON COLUMN accounts.display_name IS NULL;
COMMENT ON COLUMN accounts.phone_verified IS NULL;
COMMENT ON COLUMN accounts.phone_verified_at IS NULL;
COMMENT ON COLUMN accounts.kyc_status IS NULL;

-- +goose StatementEnd