-- +goose Up
-- +goose StatementBegin

ALTER TABLE events ADD COLUMN registration_enabled BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE events DROP COLUMN IF EXISTS registration_enabled;

-- +goose StatementEnd