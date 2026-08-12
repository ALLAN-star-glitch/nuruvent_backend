-- +goose Up
-- +goose StatementBegin

-- 1. Drop existing foreign key constraint if it exists
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS refresh_tokens_user_id_fkey;

-- 2. Rename user_id column to account_id
ALTER TABLE refresh_tokens RENAME COLUMN user_id TO account_id;

-- 3. Update the column type if needed (should be UUID already)
-- ALTER TABLE refresh_tokens ALTER COLUMN account_id TYPE UUID USING account_id::UUID;

-- 4. Add foreign key constraint to accounts table
ALTER TABLE refresh_tokens 
    ADD CONSTRAINT fk_refresh_tokens_account 
    FOREIGN KEY (account_id) 
    REFERENCES accounts(id) 
    ON DELETE CASCADE;

-- 5. Update indexes
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
CREATE INDEX idx_refresh_tokens_account_id ON refresh_tokens(account_id);

-- 6. Update primary key default if it's using old generation
ALTER TABLE refresh_tokens ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- 7. Update any existing NULL values (just in case)
UPDATE refresh_tokens SET account_id = NULL WHERE account_id IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Rollback changes
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS fk_refresh_tokens_account;
DROP INDEX IF EXISTS idx_refresh_tokens_account_id;

ALTER TABLE refresh_tokens RENAME COLUMN account_id TO user_id;

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

ALTER TABLE refresh_tokens ALTER COLUMN id SET DEFAULT uuid_generate_v4();

-- +goose StatementEnd