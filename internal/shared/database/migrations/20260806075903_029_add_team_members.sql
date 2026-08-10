-- +goose Up
-- +goose StatementBegin

-- Create team_members table
CREATE TABLE IF NOT EXISTS team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(150),
    account_id UUID NOT NULL,
    member_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'team_member',
    job_title VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_by UUID,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_team_members_account_id ON team_members (account_id);
CREATE INDEX idx_team_members_member_id ON team_members (member_id);
CREATE INDEX idx_team_members_role ON team_members (role);
CREATE INDEX idx_team_members_slug ON team_members (slug);

-- Add foreign key constraints
ALTER TABLE team_members ADD CONSTRAINT fk_team_members_account 
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE;

ALTER TABLE team_members ADD CONSTRAINT fk_team_members_member 
    FOREIGN KEY (member_id) REFERENCES accounts(id) ON DELETE CASCADE;

ALTER TABLE team_members ADD CONSTRAINT fk_team_members_creator 
    FOREIGN KEY (created_by) REFERENCES accounts(id) ON DELETE SET NULL;

-- Create update trigger for updated_at
CREATE OR REPLACE FUNCTION update_team_members_updated_at() 
RETURNS TRIGGER LANGUAGE plpgsql AS '
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
';

DROP TRIGGER IF EXISTS update_team_members_updated_at ON team_members;
CREATE TRIGGER update_team_members_updated_at
    BEFORE UPDATE ON team_members
    FOR EACH ROW
    EXECUTE FUNCTION update_team_members_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS update_team_members_updated_at ON team_members;
DROP FUNCTION IF EXISTS update_team_members_updated_at();

ALTER TABLE team_members DROP CONSTRAINT IF EXISTS fk_team_members_account;
ALTER TABLE team_members DROP CONSTRAINT IF EXISTS fk_team_members_member;
ALTER TABLE team_members DROP CONSTRAINT IF EXISTS fk_team_members_creator;

DROP TABLE IF EXISTS team_members;

-- +goose StatementEnd