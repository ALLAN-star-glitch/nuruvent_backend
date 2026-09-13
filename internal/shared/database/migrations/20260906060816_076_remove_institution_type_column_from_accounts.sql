-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- Remove redundant institution_type column
-- ============================================================

ALTER TABLE accounts DROP COLUMN IF EXISTS institution_type;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE accounts ADD COLUMN IF NOT EXISTS institution_type VARCHAR(50);

-- +goose StatementEnd