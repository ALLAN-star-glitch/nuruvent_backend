-- +goose Up
-- +goose StatementBegin

-- Drop the duplicate password column
ALTER TABLE users DROP COLUMN IF EXISTS password;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Re-add the password column (if needed for rollback)
ALTER TABLE users ADD COLUMN password VARCHAR(255) NOT NULL DEFAULT '';

-- +goose StatementEnd