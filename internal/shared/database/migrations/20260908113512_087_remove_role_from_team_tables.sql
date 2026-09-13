-- ============================================================
-- MIGRATION: Remove role from team tables
-- Description: Roles are now inherited from account level
-- ============================================================

-- +goose Up
-- +goose StatementBegin

-- Remove role column from team_invitations
ALTER TABLE team_invitations DROP COLUMN IF EXISTS role;

-- Remove role column from team_members
ALTER TABLE team_members DROP COLUMN IF EXISTS role;

-- +goose StatementEnd

-- ============================================================
-- ROLLBACK
-- ============================================================

-- +goose Down
-- +goose StatementBegin

-- Add back role column to team_invitations
ALTER TABLE team_invitations ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'trainer';

-- Add back role column to team_members
ALTER TABLE team_members ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'trainer';

-- +goose StatementEnd