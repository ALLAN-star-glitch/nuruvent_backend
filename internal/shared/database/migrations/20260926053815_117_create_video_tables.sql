-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- VIDEO CONNECTIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS video_connections (
    id                       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id                  UUID NOT NULL,
    platform                 VARCHAR(30) NOT NULL,

    external_user_id         VARCHAR(255) NOT NULL,
    external_email           VARCHAR(255) NOT NULL DEFAULT '',
    external_org_id          VARCHAR(255) NOT NULL DEFAULT '',

    access_token_encrypted   TEXT NOT NULL,
    refresh_token_encrypted  TEXT NOT NULL DEFAULT '',
    token_expires_at         TIMESTAMPTZ NOT NULL,
    scopes                   TEXT NOT NULL DEFAULT '',

    connected_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at               TIMESTAMPTZ,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One active connection per (user, platform).
CREATE UNIQUE INDEX IF NOT EXISTS idx_video_connections_active
    ON video_connections (user_id, platform)
    WHERE revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_video_connections_user
    ON video_connections (user_id);

CREATE INDEX IF NOT EXISTS idx_video_connections_platform_org
    ON video_connections (platform, external_org_id)
    WHERE revoked_at IS NULL;

-- ============================================================
-- VIDEO OAUTH STATES
-- ============================================================
CREATE TABLE IF NOT EXISTS video_oauth_states (
    state        VARCHAR(64) PRIMARY KEY,
    user_id      UUID NOT NULL,
    platform     VARCHAR(30) NOT NULL,
    return_url   TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL,
    consumed_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_video_oauth_states_expiry
    ON video_oauth_states (expires_at)
    WHERE consumed_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_video_oauth_states_user
    ON video_oauth_states (user_id);

-- ============================================================
-- VIDEO MEETINGS
-- ============================================================
CREATE TABLE IF NOT EXISTS video_meetings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL,
    platform        VARCHAR(30) NOT NULL,

    external_id     VARCHAR(255) NOT NULL,
    join_url        TEXT NOT NULL,
    start_url       TEXT NOT NULL DEFAULT '',
    password        VARCHAR(255) NOT NULL DEFAULT '',

    topic           VARCHAR(500) NOT NULL,
    start_time      TIMESTAMPTZ NOT NULL,
    duration_sec    INTEGER NOT NULL DEFAULT 0,
    timezone        VARCHAR(50) NOT NULL DEFAULT '',

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_video_meetings_external
    ON video_meetings (platform, external_id);

CREATE INDEX IF NOT EXISTS idx_video_meetings_user
    ON video_meetings (user_id);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS video_meetings;
DROP TABLE IF EXISTS video_oauth_states;
DROP TABLE IF EXISTS video_connections;
-- +goose StatementEnd