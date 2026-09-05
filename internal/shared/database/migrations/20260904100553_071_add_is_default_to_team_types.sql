-- internal/shared/database/migrations/XXX_add_is_default_to_team_types.sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE team_types ADD COLUMN IF NOT EXISTS is_default BOOLEAN DEFAULT false;

-- Update existing team types to set is_default
UPDATE team_types SET is_default = true WHERE slug = 'personal-team';
UPDATE team_types SET is_default = true WHERE slug = 'institution-team';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE team_types DROP COLUMN IF EXISTS is_default;
-- +goose StatementEnd