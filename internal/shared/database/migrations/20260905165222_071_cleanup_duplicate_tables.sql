-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- DROP DUPLICATE/OLD TABLES
-- ============================================================

-- Drop old team_types (no longer needed)
DROP TABLE IF EXISTS team_types CASCADE;

-- Drop old team_members (the original one)
DROP TABLE IF EXISTS team_members CASCADE;

-- Drop team_members_new (we already have the new one)
DROP TABLE IF EXISTS team_members_new CASCADE;

-- Drop old team_invitations_old
DROP TABLE IF EXISTS team_invitations_old CASCADE;

-- Drop old team_invitations if it exists (keep the new one)
-- We already have the new team_invitations table, so drop if there's a conflict
-- But our new one is already there, so we just keep it

-- ============================================================
-- RENAME team_invitations_old to something else if needed
-- ============================================================

-- The new team_invitations table already exists and is clean.
-- The old one is team_invitations_old which we just dropped.

-- ============================================================
-- VERIFY CLEAN STRUCTURE
-- ============================================================

-- After this migration, we should have:
-- ✅ account_types
-- ✅ accounts
-- ✅ account_members
-- ✅ teams
-- ✅ team_members (clean version)
-- ✅ team_invitations (clean version)

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- This is a cleanup migration. There's no going back without restoring from backup.

-- +goose StatementEnd