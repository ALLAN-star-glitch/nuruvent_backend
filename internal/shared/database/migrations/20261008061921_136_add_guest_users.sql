-- +goose Up

-- ============================================================
-- 1. Mark users as guests
-- ============================================================

ALTER TABLE users
  ADD COLUMN is_guest BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_users_is_guest
  ON users(is_guest)
  WHERE is_guest = TRUE;

-- ============================================================
-- 2. Backfill — create a guest user for every distinct guest
--    email in registrations that doesn't already have one.
-- ============================================================

INSERT INTO users (
  id,
  email,
  name,
  display_name,
  username,
  slug,
  password_hash,
  phone,
  account_type_id,
  email_verified,
  phone_verified,
  identity_verified,
  is_active,
  is_guest,
  created_at,
  updated_at
)
SELECT
  gen_random_uuid(),
  LOWER(r.guest_email),
  COALESCE(NULLIF(r.guest_name, ''), 'Guest'),
  COALESCE(NULLIF(r.guest_name, ''), 'Guest'),
  'guest_' || substr(md5(random()::text), 1, 12),
  'guest-' || substr(md5(random()::text), 1, 8),
  '',
  COALESCE(r.guest_phone, ''),
  (SELECT id FROM account_types WHERE slug = 'account-type-personal'),
  FALSE,
  FALSE,
  FALSE,
  TRUE,
  TRUE,
  NOW(),
  NOW()
FROM (
  SELECT DISTINCT
    LOWER(guest_email) AS guest_email,
    MIN(guest_name)    AS guest_name,
    MIN(guest_phone)   AS guest_phone
  FROM registrations
  WHERE guest_email IS NOT NULL
    AND guest_email <> ''
  GROUP BY LOWER(guest_email)
) r
WHERE NOT EXISTS (
  SELECT 1 FROM users u WHERE LOWER(u.email) = r.guest_email
);

-- ============================================================
-- 3. Link each guest registration to its user.
--
--    user_id and guest_email cannot coexist (chk_registrations_identity).
--    We choose the user row as the source of truth and clear the
--    guest_* columns on the registration.
-- ============================================================

UPDATE registrations r
SET
  user_id     = u.id,
  guest_email = NULL,
  guest_name  = NULL,
  guest_phone = NULL
FROM users u
WHERE r.user_id IS NULL
  AND r.guest_email IS NOT NULL
  AND r.guest_email <> ''
  AND LOWER(r.guest_email) = LOWER(u.email);

-- ============================================================
-- 4. Sanity check — visible in migration output
-- ============================================================

SELECT COUNT(*) AS orphan_registrations
FROM registrations
WHERE user_id IS NULL;


-- +goose Down

UPDATE registrations r
SET
  user_id     = NULL,
  guest_email = u.email,
  guest_name  = u.name,
  guest_phone = u.phone
FROM users u
WHERE r.user_id = u.id
  AND u.is_guest = TRUE;

DELETE FROM users WHERE is_guest = TRUE;

DROP INDEX IF EXISTS idx_users_is_guest;
ALTER TABLE users DROP COLUMN is_guest;