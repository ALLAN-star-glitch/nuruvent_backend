-- +goose Up
-- +goose StatementBegin

-- Migrate institution_admins to team_members if data exists
INSERT INTO team_members (
    id,
    slug,
    name,
    display_name,
    account_id,
    member_id,
    role,
    job_title,
    is_active,
    created_by,
    joined_at,
    created_at,
    updated_at
)
SELECT 
    id,
    'admin-' || id::text,
    'Admin',
    'Admin',
    institution_id,
    account_id,
    'admin',
    job_title,
    is_active,
    NULL,
    created_at,
    created_at,
    updated_at
FROM institution_admins
ON CONFLICT (id) DO NOTHING;

-- Drop institution_admins table
DROP TABLE IF EXISTS institution_admins;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Recreate institution_admins table
CREATE TABLE IF NOT EXISTS institution_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID NOT NULL,
    account_id UUID NOT NULL,
    job_title VARCHAR(100),
    is_primary BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Restore data from team_members
INSERT INTO institution_admins (
    id,
    institution_id,
    account_id,
    job_title,
    is_active,
    created_at,
    updated_at
)
SELECT 
    id,
    account_id,
    member_id,
    job_title,
    is_active,
    created_at,
    updated_at
FROM team_members 
WHERE role = 'admin';

-- +goose StatementEnd