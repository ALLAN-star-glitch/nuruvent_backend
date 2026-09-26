-- +goose Up
-- +goose StatementBegin
ALTER TABLE attendee_session_statuses
    ADD COLUMN confirmed_status VARCHAR(20);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE attendee_session_statuses
    DROP COLUMN IF EXISTS confirmed_status;
-- +goose StatementEnd