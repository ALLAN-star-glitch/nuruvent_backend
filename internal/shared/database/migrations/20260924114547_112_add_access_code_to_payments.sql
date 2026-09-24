-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments
    ADD COLUMN access_code VARCHAR(255) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments
    DROP COLUMN IF EXISTS access_code;
-- +goose StatementEnd