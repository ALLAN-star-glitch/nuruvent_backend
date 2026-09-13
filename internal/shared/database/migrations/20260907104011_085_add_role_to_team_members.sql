-- +goose Up
-- +goose StatementBegin
-- Add role column if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'team_members' AND column_name = 'role'
    ) THEN
        ALTER TABLE team_members ADD COLUMN role VARCHAR(50) NOT NULL DEFAULT 'trainer';
        CREATE INDEX idx_team_members_role ON team_members(role);
    END IF;
END $$;

-- Add invited_by column if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'team_members' AND column_name = 'invited_by'
    ) THEN
        ALTER TABLE team_members ADD COLUMN invited_by UUID REFERENCES users(id) ON DELETE SET NULL;
        CREATE INDEX idx_team_members_invited_by ON team_members(invited_by);
    END IF;
END $$;

-- Add joined_at column if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'team_members' AND column_name = 'joined_at'
    ) THEN
        ALTER TABLE team_members ADD COLUMN joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_team_members_role;
DROP INDEX IF EXISTS idx_team_members_invited_by;
ALTER TABLE team_members DROP COLUMN IF EXISTS role;
ALTER TABLE team_members DROP COLUMN IF EXISTS invited_by;
ALTER TABLE team_members DROP COLUMN IF EXISTS joined_at;
-- +goose StatementEnd