-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- ATTENDEES
-- ============================================================
CREATE TABLE IF NOT EXISTS attendees (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_type  VARCHAR(50)  NOT NULL,
    external_id    UUID         NOT NULL,
    display_name   VARCHAR(255) NOT NULL,
    email          VARCHAR(255) NOT NULL,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_attendees_external
    ON attendees (external_type, external_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_attendees_email
    ON attendees (email)
    WHERE deleted_at IS NULL;

-- ============================================================
-- SESSIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS sessions (
    id                     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_type          VARCHAR(50)  NOT NULL,
    external_id            UUID         NOT NULL,
    title                  VARCHAR(255) NOT NULL,
    scheduled_start        TIMESTAMPTZ  NOT NULL,
    scheduled_end          TIMESTAMPTZ  NOT NULL,
    duration_minutes       INTEGER      NOT NULL,
    provider               VARCHAR(20)  NOT NULL,
    provider_meeting_id    VARCHAR(255) NOT NULL DEFAULT '',
    provider_url           TEXT         NOT NULL DEFAULT '',
    status                 VARCHAR(20)  NOT NULL DEFAULT 'scheduled',
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at             TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sessions_external
    ON sessions (external_type, external_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sessions_scheduled
    ON sessions (scheduled_start)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sessions_provider
    ON sessions (provider)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sessions_meeting
    ON sessions (provider_meeting_id)
    WHERE provider_meeting_id <> '' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sessions_status
    ON sessions (status)
    WHERE deleted_at IS NULL;

-- ============================================================
-- JOIN TOKENS
-- ============================================================
CREATE TABLE IF NOT EXISTS join_tokens (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    attendee_id  UUID        NOT NULL REFERENCES attendees(id) ON DELETE CASCADE,
    session_id   UUID        NOT NULL REFERENCES sessions(id)  ON DELETE CASCADE,
    token_hash   CHAR(64)    NOT NULL UNIQUE,
    issued_at    TIMESTAMPTZ NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_join_tokens_attendee
    ON join_tokens (attendee_id);

CREATE INDEX IF NOT EXISTS idx_join_tokens_session
    ON join_tokens (session_id);

CREATE INDEX IF NOT EXISTS idx_join_tokens_expires
    ON join_tokens (expires_at)
    WHERE revoked_at IS NULL;

-- ============================================================
-- ATTENDANCE RECORDS
-- ============================================================
CREATE TABLE IF NOT EXISTS attendance_records (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    attendee_id       UUID        NOT NULL REFERENCES attendees(id) ON DELETE CASCADE,
    session_id        UUID        NOT NULL REFERENCES sessions(id)  ON DELETE CASCADE,
    join_time         TIMESTAMPTZ NOT NULL,
    leave_time        TIMESTAMPTZ,
    duration_seconds  BIGINT      NOT NULL DEFAULT 0,
    source            VARCHAR(30) NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_attendance_records_attendee_session
    ON attendance_records (attendee_id, session_id);

CREATE INDEX IF NOT EXISTS idx_attendance_records_join_time
    ON attendance_records (join_time);

CREATE INDEX IF NOT EXISTS idx_attendance_records_open
    ON attendance_records (session_id)
    WHERE leave_time IS NULL;

-- ============================================================
-- ATTENDEE SESSION STATUS (derived, denormalized)
-- ============================================================
CREATE TABLE IF NOT EXISTS attendee_session_statuses (
    attendee_id            UUID        NOT NULL REFERENCES attendees(id) ON DELETE CASCADE,
    session_id             UUID        NOT NULL REFERENCES sessions(id)  ON DELETE CASCADE,
    derived_status         VARCHAR(20) NOT NULL,
    total_duration_seconds BIGINT      NOT NULL DEFAULT 0,
    host_confirmed         BOOLEAN     NOT NULL DEFAULT false,
    confirmed_by           UUID,
    confirmed_at           TIMESTAMPTZ,
    confirm_reason         TEXT        NOT NULL DEFAULT '',
    last_derived_at        TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (attendee_id, session_id)
);

CREATE INDEX IF NOT EXISTS idx_ass_status
    ON attendee_session_statuses (derived_status);

CREATE INDEX IF NOT EXISTS idx_ass_confirmed
    ON attendee_session_statuses (host_confirmed)
    WHERE host_confirmed = true;

-- ============================================================
-- ATTENDEE EVENT STATUS (roll-up, denormalized)
-- ============================================================
CREATE TABLE IF NOT EXISTS attendee_event_statuses (
    attendee_id            UUID        NOT NULL REFERENCES attendees(id) ON DELETE CASCADE,
    external_type          VARCHAR(50) NOT NULL,
    external_id            UUID        NOT NULL,
    derived_status         VARCHAR(20) NOT NULL,
    sessions_total         INTEGER     NOT NULL DEFAULT 0,
    sessions_attended      INTEGER     NOT NULL DEFAULT 0,
    sessions_confirmed     INTEGER     NOT NULL DEFAULT 0,
    total_duration_seconds BIGINT      NOT NULL DEFAULT 0,
    last_derived_at        TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (attendee_id, external_type, external_id)
);

CREATE INDEX IF NOT EXISTS idx_aes_status
    ON attendee_event_statuses (derived_status);

CREATE INDEX IF NOT EXISTS idx_aes_external
    ON attendee_event_statuses (external_type, external_id);

-- ============================================================
-- ATTENDANCE OVERRIDES (append-only audit)
-- ============================================================
CREATE TABLE IF NOT EXISTS attendance_overrides (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    attendee_id  UUID        NOT NULL REFERENCES attendees(id) ON DELETE CASCADE,
    session_id   UUID        NOT NULL REFERENCES sessions(id)  ON DELETE CASCADE,
    actor_id     UUID        NOT NULL,
    prior_status VARCHAR(20) NOT NULL,
    new_status   VARCHAR(20) NOT NULL,
    reason       TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_overrides_attendee_session
    ON attendance_overrides (attendee_id, session_id);

CREATE INDEX IF NOT EXISTS idx_overrides_actor
    ON attendance_overrides (actor_id);

CREATE INDEX IF NOT EXISTS idx_overrides_created
    ON attendance_overrides (created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS attendance_overrides;
DROP TABLE IF EXISTS attendee_event_statuses;
DROP TABLE IF EXISTS attendee_session_statuses;
DROP TABLE IF EXISTS attendance_records;
DROP TABLE IF EXISTS join_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS attendees;
-- +goose StatementEnd