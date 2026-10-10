-- +goose Up
-- Fix: allow 'learner' in team_invitations.role
--
-- The original constraint predates the learner role and only allowed
-- 'account_admin' and 'trainer'. Invites with role='learner' were being
-- rejected at the DB level (SQLSTATE 23514).

ALTER TABLE team_invitations
  DROP CONSTRAINT IF EXISTS team_invitations_role_check;

ALTER TABLE team_invitations
  ADD CONSTRAINT team_invitations_role_check
  CHECK (role IN ('account_admin', 'trainer', 'learner'));

-- +goose Down
ALTER TABLE team_invitations
  DROP CONSTRAINT IF EXISTS team_invitations_role_check;

ALTER TABLE team_invitations
  ADD CONSTRAINT team_invitations_role_check
  CHECK (role IN ('account_admin', 'trainer'));