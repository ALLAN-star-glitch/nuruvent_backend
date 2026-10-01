-- +goose Up

-- ============================================================
-- join_tokens.redeemed_at
-- ============================================================
--
-- Records the first time an attendee opens their personalized join
-- link. NULL means the token was issued but never used.
--
-- The attendance fetch matcher queries (session_id, redeemed_at) to
-- bind Meet participants to attendees by token redemption instead of
-- by display name. The Meet Reports API does not expose participant
-- emails, so display-name matching fails whenever the participant's
-- Google name differs from their registration name — which is the
-- common case.
--
-- See: internal/modules/attendance/docs/join-token-flow.md

ALTER TABLE join_tokens
    ADD COLUMN IF NOT EXISTS redeemed_at TIMESTAMPTZ;

-- Partial composite index — the matcher's query always filters on
-- redeemed_at IS NOT NULL, so indexing only redeemed rows keeps the
-- index small and fast.
CREATE INDEX IF NOT EXISTS idx_join_tokens_session_redeemed
    ON join_tokens (session_id, redeemed_at)
    WHERE redeemed_at IS NOT NULL;

-- +goose Down

DROP INDEX IF EXISTS idx_join_tokens_session_redeemed;

ALTER TABLE join_tokens
    DROP COLUMN IF EXISTS redeemed_at;