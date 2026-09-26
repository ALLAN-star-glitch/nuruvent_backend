-- +goose Up
-- +goose StatementBegin
ALTER TABLE sessions
    ADD COLUMN provider_session_id VARCHAR(255) NOT NULL DEFAULT '';

-- The uniqueness of a session is now (external_type, external_id,
-- provider_session_id). The old index on just (external_type,
-- external_id) still exists and remains useful for parent-scoped
-- queries.
CREATE INDEX IF NOT EXISTS idx_sessions_provider_session
    ON sessions (provider_session_id)
    WHERE provider_session_id <> '' AND deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_sessions_provider_session;
ALTER TABLE sessions DROP COLUMN IF EXISTS provider_session_id;
-- +goose StatementEnd