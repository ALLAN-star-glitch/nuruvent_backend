-- +goose Up
-- +goose StatementBegin

-- 1. Add the column with a safe default
ALTER TABLE sessions 
  ADD COLUMN is_manual_link boolean NOT NULL DEFAULT false;

-- 2. Backfill existing records.
-- If a session has a provider URL but its meeting ID does NOT exist in
-- video_meetings, the host pasted the link manually.
UPDATE sessions
SET is_manual_link = true
WHERE provider_meeting_id IS NOT NULL 
  AND provider_meeting_id != ''
  AND provider_url IS NOT NULL
  AND provider_url != ''
  AND provider_meeting_id NOT IN (
      SELECT external_id FROM video_meetings
  );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE sessions 
  DROP COLUMN IF EXISTS is_manual_link;

-- +goose StatementEnd