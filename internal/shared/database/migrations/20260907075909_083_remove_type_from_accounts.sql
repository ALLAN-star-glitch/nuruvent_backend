-- +goose Up
-- +goose StatementBegin
-- Remove type column from accounts table
DROP INDEX IF EXISTS idx_accounts_type;
ALTER TABLE accounts DROP COLUMN IF EXISTS type;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Restore type column (in case of rollback)
ALTER TABLE accounts ADD COLUMN type VARCHAR(50) DEFAULT 'personal';
CREATE INDEX idx_accounts_type ON accounts(type);
UPDATE accounts SET type = 'personal' WHERE type IS NULL;
ALTER TABLE accounts ALTER COLUMN type SET NOT NULL;
-- +goose StatementEnd