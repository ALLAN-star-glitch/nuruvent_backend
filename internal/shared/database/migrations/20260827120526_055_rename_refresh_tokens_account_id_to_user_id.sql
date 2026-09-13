-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- MIGRATION: Rename account_id to user_id in refresh_tokens
-- ============================================================

-- Rename the column
ALTER TABLE refresh_tokens RENAME COLUMN account_id TO user_id;

-- Rename the index
ALTER INDEX idx_refresh_tokens_account_id RENAME TO idx_refresh_tokens_user_id;

-- Rename the foreign key constraint
ALTER TABLE refresh_tokens RENAME CONSTRAINT fk_refresh_tokens_account TO fk_refresh_tokens_user;

-- Update comment
COMMENT ON COLUMN refresh_tokens.user_id IS 'References the user who owns this refresh token';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Rename back
ALTER TABLE refresh_tokens RENAME COLUMN user_id TO account_id;

-- Rename index back
ALTER INDEX idx_refresh_tokens_user_id RENAME TO idx_refresh_tokens_account_id;

-- Rename foreign key back
ALTER TABLE refresh_tokens RENAME CONSTRAINT fk_refresh_tokens_user TO fk_refresh_tokens_account;

-- +goose StatementEnd