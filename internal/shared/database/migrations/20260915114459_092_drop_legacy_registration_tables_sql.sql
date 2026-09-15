-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS event_registrations CASCADE;
DROP TABLE IF EXISTS attendees CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Intentionally non-reversible. Restore from a backup if needed.
-- The original DDL is preserved in git history (migrations 005 and 006).
SELECT 1;
-- +goose StatementEnd