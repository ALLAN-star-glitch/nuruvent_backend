-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- MIGRATION: Remove generic casbin policies (account:xxx format)
-- ============================================================

-- Remove policies with account:xxx domain format
DELETE FROM casbin_rule 
WHERE ptype = 'p' 
AND v1 LIKE 'account:%';

-- Remove grouping policies with account:xxx domain format
DELETE FROM casbin_rule 
WHERE ptype = 'g' 
AND v2 LIKE 'account:%';

-- ============================================================
-- LOG: Show remaining policies (for debugging)
-- ============================================================
DO $$
DECLARE
    policy_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO policy_count FROM casbin_rule WHERE ptype = 'p';
    RAISE NOTICE 'Remaining policy rules: %', policy_count;
    
    SELECT COUNT(*) INTO policy_count FROM casbin_rule WHERE ptype = 'g';
    RAISE NOTICE 'Remaining grouping policies: %', policy_count;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- ============================================================
-- DOWN: Restore policies (can't restore deleted data)
-- ============================================================

-- WARNING: This migration cannot be fully reversed
-- The deleted policies cannot be restored automatically

DO $$
BEGIN
    RAISE NOTICE 'WARNING: Deleted policies cannot be restored automatically.';
    RAISE NOTICE 'Please re-run seeders to restore policies.';
END $$;

-- +goose StatementEnd