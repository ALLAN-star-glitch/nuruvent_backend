-- +goose Up

-- Backfill account_members for guests whose account creation
-- succeeded but whose membership insert failed.

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
    SELECT 1
    FROM account_members am
    WHERE am.user_id = u.id
      AND am.account_id = a.id
      AND am.deleted_at IS NULL
  );

-- Backfill Casbin grouping rules for the same users.
-- Note: casbin_rule.v0 is varchar, users.id is uuid —
-- explicit ::text casts are required.

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
WHERE u.deleted_at IS NULL
  AND a.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM account_members am
    WHERE am.user_id = u.id
      AND am.account_id = a.id
      AND am.deleted_at IS NULL
  )
  AND NOT EXISTS (
    SELECT 1
    FROM casbin_rule cr
    WHERE cr.ptype = 'g'
      AND cr.v0 = u.id::text
      AND cr.v1 = 'account_admin'
      AND cr.v2 = 'account:' || a.id::text
  );

-- +goose Down

-- Rollback is a no-op. Deleting the backfilled rows would be
-- indistinguishable from legitimate memberships created later.

SELECT 1;