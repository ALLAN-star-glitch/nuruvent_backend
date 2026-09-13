-- +goose Up
ALTER TABLE team_invitations
  ADD COLUMN role VARCHAR(50);

-- Backfill: existing invitations default to 'trainer'.
-- In dev, you can safely delete old invitations instead of backfilling.
UPDATE team_invitations
SET role = 'trainer'
WHERE role IS NULL;

ALTER TABLE team_invitations
  ALTER COLUMN role SET NOT NULL;

ALTER TABLE team_invitations
  ADD CONSTRAINT team_invitations_role_check
  CHECK (role IN ('account_admin', 'trainer'));

-- +goose Down
ALTER TABLE team_invitations DROP CONSTRAINT IF EXISTS team_invitations_role_check;
ALTER TABLE team_invitations DROP COLUMN IF EXISTS role;