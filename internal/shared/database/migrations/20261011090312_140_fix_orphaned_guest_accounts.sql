-- +goose Up

-- ============================================================
-- Step 1: create missing accounts
-- ============================================================
-- Guests whose account insert was rolled back have no account
-- row. Recreate them from the user's stored identity so the
-- membership and role steps below have something to attach to.
-- ============================================================

INSERT INTO accounts (
  id,
  name,
  display_name,
  slug,
  email,
  phone,
  account_type_id,
  status,
  created_by,
  created_at,
  updated_at
)
SELECT
  gen_random_uuid(),
  u.name || '''s Account',
  u.name,
  'acct-' || substr(md5(random()::text || clock_timestamp()::text), 1, 12),
  u.email,
  COALESCE(u.phone, ''),
  u.account_type_id,
  'active',
  u.id,
  now(),
  now()
FROM users u
WHERE u.is_guest = true
  AND u.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM accounts a
    WHERE a.created_by = u.id
      AND a.deleted_at IS NULL
  );

-- ============================================================
-- Step 2: create the missing membership
-- ============================================================

INSERT INTO account_members (
  id,
  account_id,
  user_id,
  role,
  is_active,
  invited_by,
  joined_at,
  created_at,
  updated_at
)
SELECT
  gen_random_uuid(),
  a.id,
  u.id,
  'account_admin',
  true,
  NULL,
  COALESCE(a.created_at, now()),
  now(),
  now()
FROM users u
JOIN accounts a ON a.created_by = u.id
WHERE u.deleted_at IS NULL
  AND a.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM account_members am
    WHERE am.user_id = u.id
      AND am.account_id = a.id
      AND am.deleted_at IS NULL
  );

-- ============================================================
-- Step 3: grant the Casbin role
-- ============================================================

INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
SELECT
  'g',
  u.id::text,
  'account_admin',
  'account:' || a.id::text,
  '',
  '',
  ''
FROM users u
JOIN accounts a ON a.created_by = u.id
JOIN account_members am
  ON am.user_id = u.id
 AND am.account_id = a.id
 AND am.deleted_at IS NULL
WHERE u.deleted_at IS NULL
  AND a.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM casbin_rule cr
    WHERE cr.ptype = 'g'
      AND cr.v0 = u.id::text
      AND cr.v1 = 'account_admin'
      AND cr.v2 = 'account:' || a.id::text
  );

-- +goose Down

SELECT 1;