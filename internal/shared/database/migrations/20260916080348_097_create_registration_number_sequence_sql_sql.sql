-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE IF NOT EXISTS registration_number_seq START 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS registration_number_seq;
-- +goose StatementEnd