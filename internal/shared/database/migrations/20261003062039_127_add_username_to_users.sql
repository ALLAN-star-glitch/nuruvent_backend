-- +goose Up
-- Add username column to users.
--
-- Username is a stable, unique, immutable identity key derived from the
-- user's email local part. It is never shown in the UI and cannot be
-- changed by the user. Used for cross-module identity matching where
-- display_name is unsafe (e.g. matching anonymous Zoom participants to
-- attendees).

ALTER TABLE users ADD COLUMN username text;

-- Backfill from email local part, normalized to lowercase alphanumerics.
-- For users whose email local part is already unique (the common case),
-- this produces a clean username with no collision handling needed.
UPDATE users
SET username = lower(regexp_replace(split_part(email, '@', 1), '[^a-z0-9]', '', 'g'))
WHERE username IS NULL;

ALTER TABLE users ALTER COLUMN username SET NOT NULL;

ALTER TABLE users ADD CONSTRAINT users_username_unique UNIQUE (username);

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_unique;
ALTER TABLE users DROP COLUMN IF EXISTS username;