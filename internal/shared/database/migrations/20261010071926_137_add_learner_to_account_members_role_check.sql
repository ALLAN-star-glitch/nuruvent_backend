-- +goose Up

-- Extend the account_members.role CHECK constraint to include 'learner'.

ALTER TABLE account_members
    DROP CONSTRAINT IF EXISTS account_members_role_check;

ALTER TABLE account_members
    ADD CONSTRAINT account_members_role_check
    CHECK (role::text = ANY (ARRAY['account_admin'::character varying, 'trainer'::character varying, 'learner'::character varying]::text[]));

-- +goose Down

ALTER TABLE account_members
    DROP CONSTRAINT IF EXISTS account_members_role_check;

ALTER TABLE account_members
    ADD CONSTRAINT account_members_role_check
    CHECK (role::text = ANY (ARRAY['account_admin'::character varying, 'trainer'::character varying]::text[]));