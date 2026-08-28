-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- MIGRATION: Make institution_id nullable for personal teams
-- ============================================================

ALTER TABLE team_members ALTER COLUMN institution_id DROP NOT NULL;

COMMENT ON COLUMN team_members.institution_id IS 'The institution this member belongs to (NULL for personal teams)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE team_members ALTER COLUMN institution_id SET NOT NULL;

-- +goose StatementEnd