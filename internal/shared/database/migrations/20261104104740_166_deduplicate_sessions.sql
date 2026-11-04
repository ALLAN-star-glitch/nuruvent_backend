-- +goose Up
-- +goose StatementBegin

-- 1. Delete duplicate sessions, keeping only the newest one per (event, title).
DELETE FROM sessions s1
USING sessions s2
WHERE s1.id > s2.id
  AND s1.external_type = s2.external_type
  AND s1.external_id = s2.external_id
  AND s1.title = s2.title;

-- 2. Add the unique constraint to prevent future duplicates.
ALTER TABLE sessions 
ADD CONSTRAINT unique_session_per_event 
UNIQUE (external_type, external_id, title);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS unique_session_per_event;
-- +goose StatementEnd