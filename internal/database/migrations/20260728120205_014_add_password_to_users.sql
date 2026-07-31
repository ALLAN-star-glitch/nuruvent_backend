-- +goose Up
-- +goose StatementBegin

-- Add password column to users table
ALTER TABLE users ADD COLUMN password VARCHAR(255) NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove password column from users table
ALTER TABLE users DROP COLUMN IF EXISTS password;

-- +goose StatementEnd