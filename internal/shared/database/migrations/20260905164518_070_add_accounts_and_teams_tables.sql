-- +goose Up
-- +goose StatementBegin

-- 1. Create accounts table
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(50),
    type VARCHAR(50) NOT NULL CHECK (type IN ('personal', 'institution')),
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'inactive')),
    logo_url VARCHAR(500),
    website VARCHAR(255),
    description TEXT,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_accounts_type ON accounts(type);
CREATE INDEX IF NOT EXISTS idx_accounts_slug ON accounts(slug);
CREATE INDEX IF NOT EXISTS idx_accounts_email ON accounts(email);
CREATE INDEX IF NOT EXISTS idx_accounts_deleted_at ON accounts(deleted_at);
CREATE INDEX IF NOT EXISTS idx_accounts_status ON accounts(status);

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_accounts_created_by' AND table_name = 'accounts'
    ) THEN
        ALTER TABLE accounts 
            ADD CONSTRAINT fk_accounts_created_by 
            FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- 2. Create account_members table
CREATE TABLE IF NOT EXISTS account_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('account_admin', 'trainer')),
    is_active BOOLEAN DEFAULT true,
    invited_by UUID,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(account_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_account_members_account_id ON account_members(account_id);
CREATE INDEX IF NOT EXISTS idx_account_members_user_id ON account_members(user_id);
CREATE INDEX IF NOT EXISTS idx_account_members_role ON account_members(role);
CREATE INDEX IF NOT EXISTS idx_account_members_deleted_at ON account_members(deleted_at);
CREATE INDEX IF NOT EXISTS idx_account_members_is_active ON account_members(is_active);

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_account_members_account' AND table_name = 'account_members'
    ) THEN
        ALTER TABLE account_members 
            ADD CONSTRAINT fk_account_members_account 
            FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_account_members_user' AND table_name = 'account_members'
    ) THEN
        ALTER TABLE account_members 
            ADD CONSTRAINT fk_account_members_user 
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_account_members_invited_by' AND table_name = 'account_members'
    ) THEN
        ALTER TABLE account_members 
            ADD CONSTRAINT fk_account_members_invited_by 
            FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- 3. Create teams table
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('personal', 'institution')),
    description TEXT,
    logo_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(account_id, name)
);

CREATE INDEX IF NOT EXISTS idx_teams_account_id ON teams(account_id);
CREATE INDEX IF NOT EXISTS idx_teams_type ON teams(type);
CREATE INDEX IF NOT EXISTS idx_teams_slug ON teams(slug);
CREATE INDEX IF NOT EXISTS idx_teams_deleted_at ON teams(deleted_at);
CREATE INDEX IF NOT EXISTS idx_teams_is_active ON teams(is_active);

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_teams_account' AND table_name = 'teams'
    ) THEN
        ALTER TABLE teams 
            ADD CONSTRAINT fk_teams_account 
            FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_teams_created_by' AND table_name = 'teams'
    ) THEN
        ALTER TABLE teams 
            ADD CONSTRAINT fk_teams_created_by 
            FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- 4. Create team_members_new table
CREATE TABLE IF NOT EXISTS team_members_new (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('account_admin', 'trainer')),
    is_active BOOLEAN DEFAULT true,
    invited_by UUID,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(team_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_team_members_new_team_id ON team_members_new(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_new_user_id ON team_members_new(user_id);
CREATE INDEX IF NOT EXISTS idx_team_members_new_role ON team_members_new(role);
CREATE INDEX IF NOT EXISTS idx_team_members_new_deleted_at ON team_members_new(deleted_at);
CREATE INDEX IF NOT EXISTS idx_team_members_new_is_active ON team_members_new(is_active);

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_team_members_new_team' AND table_name = 'team_members_new'
    ) THEN
        ALTER TABLE team_members_new 
            ADD CONSTRAINT fk_team_members_new_team 
            FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_team_members_new_user' AND table_name = 'team_members_new'
    ) THEN
        ALTER TABLE team_members_new 
            ADD CONSTRAINT fk_team_members_new_user 
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_team_members_new_invited_by' AND table_name = 'team_members_new'
    ) THEN
        ALTER TABLE team_members_new 
            ADD CONSTRAINT fk_team_members_new_invited_by 
            FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- 5. Handle existing team_invitations
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'team_invitations'
    ) THEN
        EXECUTE 'ALTER TABLE team_invitations RENAME TO team_invitations_old';
    END IF;
END $$;

-- 6. Create team_invitations table
CREATE TABLE IF NOT EXISTS team_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('account_admin', 'trainer')),
    token VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'expired')),
    invited_by UUID,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE,
    declined_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_team_invitations_team_id ON team_invitations(team_id);
CREATE INDEX IF NOT EXISTS idx_team_invitations_email ON team_invitations(email);
CREATE INDEX IF NOT EXISTS idx_team_invitations_token ON team_invitations(token);
CREATE INDEX IF NOT EXISTS idx_team_invitations_status ON team_invitations(status);
CREATE INDEX IF NOT EXISTS idx_team_invitations_deleted_at ON team_invitations(deleted_at);

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_team_invitations_team' AND table_name = 'team_invitations'
    ) THEN
        ALTER TABLE team_invitations 
            ADD CONSTRAINT fk_team_invitations_team 
            FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_team_invitations_invited_by' AND table_name = 'team_invitations'
    ) THEN
        ALTER TABLE team_invitations 
            ADD CONSTRAINT fk_team_invitations_invited_by 
            FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- 7. Add account_type_id to users
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'account_type_id'
    ) THEN
        ALTER TABLE users ADD COLUMN account_type_id UUID;
    END IF;
END $$;

-- 8. Add foreign key for account_type_id
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'account_types'
    ) THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints 
            WHERE constraint_name = 'fk_users_account_type' AND table_name = 'users'
        ) THEN
            ALTER TABLE users 
                ADD CONSTRAINT fk_users_account_type 
                FOREIGN KEY (account_type_id) REFERENCES account_types(id) ON DELETE SET NULL;
        END IF;
    END IF;
END $$;

-- 9. Skip data migration - we'll handle it separately
-- The existing team_members table will remain as is for now

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Rollback: Drop new tables
DROP TABLE IF EXISTS team_invitations CASCADE;
DROP TABLE IF EXISTS team_members_new CASCADE;
DROP TABLE IF EXISTS teams CASCADE;
DROP TABLE IF EXISTS account_members CASCADE;
DROP TABLE IF EXISTS accounts CASCADE;

-- Restore old tables if they were renamed
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'team_invitations_old'
    ) THEN
        EXECUTE 'ALTER TABLE team_invitations_old RENAME TO team_invitations';
    END IF;
END $$;

-- Remove column from users
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'account_type_id'
    ) THEN
        ALTER TABLE users DROP COLUMN IF EXISTS account_type_id;
    END IF;
END $$;

-- +goose StatementEnd