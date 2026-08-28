-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- MIGRATION: Remove generic Casbin policies with '*' domain
-- Description: Fix security issue where event_manager can access
--              ANY account due to generic role hierarchy with '*'
-- ============================================================

-- ============================================================
-- 1. DELETE GENERIC ROLE HIERARCHY WITH '*' DOMAIN
--    Line 54: g event_manager team_member *
--    This gives event_manager team_member permissions across ALL accounts
-- ============================================================
DELETE FROM casbin_rule 
WHERE ptype = 'g' 
AND v0 = 'event_manager' 
AND v1 = 'team_member' 
AND v2 = '*';

-- ============================================================
-- 2. DELETE GENERIC POLICY RULES WITH '*' DOMAIN (if any exist)
--    These give access to ALL accounts
-- ============================================================
DELETE FROM casbin_rule 
WHERE ptype = 'p' 
AND v2 = '*';

-- ============================================================
-- 3. VERIFY - Check if any wildcard domains remain
-- ============================================================
DO $$
DECLARE
    generic_policies INTEGER;
    generic_hierarchy INTEGER;
BEGIN
    SELECT COUNT(*) INTO generic_policies 
    FROM casbin_rule 
    WHERE ptype = 'p' AND v2 = '*';
    
    IF generic_policies > 0 THEN
        RAISE NOTICE '⚠️  WARNING: % generic policies with ''*'' domain still exist!', generic_policies;
    ELSE
        RAISE NOTICE '✅ All generic policies with ''*'' domain removed successfully!';
    END IF;

    SELECT COUNT(*) INTO generic_hierarchy 
    FROM casbin_rule 
    WHERE ptype = 'g' AND v2 = '*';
    
    IF generic_hierarchy > 0 THEN
        RAISE NOTICE '⚠️  WARNING: % generic role hierarchies with ''*'' domain still exist!', generic_hierarchy;
    ELSE
        RAISE NOTICE '✅ All generic role hierarchies with ''*'' domain removed successfully!';
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- ============================================================
-- DOWN: Restore generic policies (rollback - SECURITY RISK)
-- ============================================================

-- 1. Restore generic role hierarchy (SECURITY RISK)
INSERT INTO casbin_rule (ptype, v0, v1, v2) 
SELECT 'g', 'event_manager', 'team_member', '*'
WHERE NOT EXISTS (
    SELECT 1 FROM casbin_rule 
    WHERE ptype = 'g' 
    AND v0 = 'event_manager' 
    AND v1 = 'team_member' 
    AND v2 = '*'
);

-- 2. Restore super_admin wildcard policy (SECURITY RISK)
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) 
SELECT 'p', 'super_admin', 'platform', '*', '*'
WHERE NOT EXISTS (
    SELECT 1 FROM casbin_rule 
    WHERE ptype = 'p' 
    AND v0 = 'super_admin' 
    AND v1 = 'platform' 
    AND v2 = '*' 
    AND v3 = '*'
);

-- +goose StatementEnd