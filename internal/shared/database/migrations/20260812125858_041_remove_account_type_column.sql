-- +goose Up
-- +goose StatementBegin
ALTER TABLE accounts DROP COLUMN IF EXISTS account_type;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts ADD COLUMN account_type VARCHAR(50);
-- +goose StatementEnd