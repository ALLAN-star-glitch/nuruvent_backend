-- +goose Up
-- +goose StatementBegin
-- Remove role column from team_members table
ALTER TABLE team_members DROP COLUMN IF EXISTS role;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Restore role column (if rollback needed)
ALTER TABLE team_members ADD COLUMN role VARCHAR(50);
-- +goose StatementEnd