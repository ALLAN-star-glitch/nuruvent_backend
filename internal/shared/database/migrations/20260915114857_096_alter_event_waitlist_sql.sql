-- +goose Up
-- +goose StatementBegin
ALTER TABLE event_waitlist
    ADD COLUMN guest_email VARCHAR(255),
    ADD COLUMN guest_name  VARCHAR(255),
    ALTER COLUMN user_id DROP NOT NULL;

ALTER TABLE event_waitlist
    ADD CONSTRAINT chk_event_waitlist_identity CHECK (
        (user_id IS NOT NULL AND guest_email IS NULL) OR
        (user_id IS NULL AND guest_email IS NOT NULL)
    );

CREATE INDEX idx_event_waitlist_guest_email ON event_waitlist(guest_email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_event_waitlist_guest_email;

ALTER TABLE event_waitlist
    DROP CONSTRAINT IF EXISTS chk_event_waitlist_identity;

ALTER TABLE event_waitlist
    DROP COLUMN IF EXISTS guest_name,
    DROP COLUMN IF EXISTS guest_email;

ALTER TABLE event_waitlist
    ALTER COLUMN user_id SET NOT NULL;
-- +goose StatementEnd