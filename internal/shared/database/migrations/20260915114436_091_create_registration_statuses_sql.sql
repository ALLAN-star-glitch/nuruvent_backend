-- +goose Up
-- +goose StatementBegin
CREATE TABLE registration_statuses (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug         VARCHAR(50)  NOT NULL UNIQUE,
    name         VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description  TEXT,
    color        VARCHAR(20),
    icon         VARCHAR(50),
    sort_order   INTEGER NOT NULL DEFAULT 0,
    is_final     BOOLEAN NOT NULL DEFAULT FALSE,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_registration_statuses_slug       ON registration_statuses(slug);
CREATE INDEX idx_registration_statuses_is_active  ON registration_statuses(is_active);
CREATE INDEX idx_registration_statuses_sort_order ON registration_statuses(sort_order);

CREATE OR REPLACE FUNCTION update_registration_statuses_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_registration_statuses_updated_at
    BEFORE UPDATE ON registration_statuses
    FOR EACH ROW EXECUTE FUNCTION update_registration_statuses_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_registration_statuses_updated_at ON registration_statuses;
DROP FUNCTION IF EXISTS update_registration_statuses_updated_at();
DROP TABLE IF EXISTS registration_statuses CASCADE;
-- +goose StatementEnd