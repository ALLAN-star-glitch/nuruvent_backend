-- +goose Up
ALTER TABLE attendees ADD COLUMN google_meet_user_id varchar(255) NOT NULL DEFAULT '';

CREATE INDEX idx_attendees_gmeet_user_id
    ON attendees (google_meet_user_id)
    WHERE google_meet_user_id <> '' AND deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_attendees_gmeet_user_id;
ALTER TABLE attendees DROP COLUMN IF EXISTS google_meet_user_id;