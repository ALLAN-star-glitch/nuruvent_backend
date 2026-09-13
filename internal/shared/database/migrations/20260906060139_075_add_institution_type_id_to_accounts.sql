-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- STEP 1: Add institution_type_id to accounts
-- ============================================================

ALTER TABLE accounts ADD COLUMN IF NOT EXISTS institution_type_id UUID;

-- ============================================================
-- STEP 2: Add foreign key constraint
-- ============================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.constraint_column_usage 
        WHERE constraint_name = 'fk_accounts_institution_type' 
        AND table_name = 'accounts'
    ) THEN
        ALTER TABLE accounts ADD CONSTRAINT fk_accounts_institution_type 
            FOREIGN KEY (institution_type_id) REFERENCES institution_types(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ============================================================
-- STEP 3: Add index for performance
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_accounts_institution_type_id ON accounts(institution_type_id);

-- ============================================================
-- STEP 4: Verify migration
-- ============================================================

DO $$
DECLARE
    accounts_with_type INTEGER;
    total_accounts INTEGER;
BEGIN
    SELECT COUNT(*) INTO total_accounts FROM accounts WHERE type = 'institution';
    SELECT COUNT(*) INTO accounts_with_type FROM accounts WHERE type = 'institution' AND institution_type_id IS NOT NULL;
    
    IF total_accounts > 0 THEN
        RAISE NOTICE '✅ % institution accounts have institution_type_id set', accounts_with_type;
        RAISE NOTICE '⚠️  % institution accounts need institution_type_id', total_accounts - accounts_with_type;
    ELSE
        RAISE NOTICE 'ℹ️  No institution accounts found';
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE accounts DROP CONSTRAINT IF EXISTS fk_accounts_institution_type;
ALTER TABLE accounts DROP COLUMN IF EXISTS institution_type_id;
DROP INDEX IF EXISTS idx_accounts_institution_type_id;

-- +goose StatementEnd