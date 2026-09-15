-- +goose Up
-- +goose StatementBegin
CREATE TABLE event_registrations (
    registration_id UUID PRIMARY KEY REFERENCES registrations(id) ON DELETE CASCADE,
    event_id        UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_event_registrations_event_id ON event_registrations(event_id);
CREATE INDEX idx_event_registrations_user_id  ON event_registrations(user_id);

CREATE UNIQUE INDEX uniq_event_registrations_active
    ON event_registrations (event_id, user_id)
    WHERE user_id IS NOT NULL;

CREATE OR REPLACE FUNCTION update_event_registrations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_event_registrations_updated_at
    BEFORE UPDATE ON event_registrations
    FOR EACH ROW EXECUTE FUNCTION update_event_registrations_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_event_registrations_updated_at ON event_registrations;
DROP FUNCTION IF EXISTS update_event_registrations_updated_at();
DROP TABLE IF EXISTS event_registrations CASCADE;
-- +goose StatementEnd