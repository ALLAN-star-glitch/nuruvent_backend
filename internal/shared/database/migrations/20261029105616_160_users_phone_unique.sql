-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX users_phone_unique
    ON users (phone)
    WHERE phone IS NOT NULL
      AND phone <> ''
      AND deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users_phone_unique;
-- +goose StatementEnd