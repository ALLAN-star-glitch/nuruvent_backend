-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS attendee_event_statuses
    RENAME TO attendee_rollup_statuses;

ALTER INDEX IF EXISTS idx_aes_status
    RENAME TO idx_ars_status;

ALTER INDEX IF EXISTS idx_aes_external
    RENAME TO idx_ars_external;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER INDEX IF EXISTS idx_ars_status
    RENAME TO idx_aes_status;

ALTER INDEX IF EXISTS idx_ars_external
    RENAME TO idx_aes_external;

ALTER TABLE IF EXISTS attendee_rollup_statuses
    RENAME TO attendee_event_statuses;
-- +goose StatementEnd