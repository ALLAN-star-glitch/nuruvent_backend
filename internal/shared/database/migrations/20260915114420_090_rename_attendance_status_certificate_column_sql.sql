-- +goose Up
-- +goose StatementBegin
ALTER TABLE attendance_statuses
    RENAME COLUMN can_issue_certificate TO counts_as_attended;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE attendance_statuses
    RENAME COLUMN counts_as_attended TO can_issue_certificate;
-- +goose StatementEnd