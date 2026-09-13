-- +goose Up
-- +goose StatementBegin
-- ============================================================
-- MIGRATION: Final team_members schema
-- ============================================================

-- ============================================================
-- 1. REMOVE UNNECESSARY COLUMNS
-- ============================================================
ALTER TABLE team_members 
DROP COLUMN IF EXISTS slug,
DROP COLUMN IF EXISTS name,
DROP COLUMN IF EXISTS display_name,
DROP COLUMN IF EXISTS member_id,
DROP COLUMN IF EXISTS role,
DROP COLUMN IF EXISTS job_title;

-- ============================================================
-- 2. RENAME user_id TO member_id (for clarity)
-- ============================================================
ALTER TABLE team_members RENAME COLUMN user_id TO member_id;

-- ============================================================
-- 3. ADD institution_id (NULLABLE for personal teams!)
-- ============================================================
ALTER TABLE team_members 
ADD COLUMN institution_id UUID REFERENCES institutions(id);

-- ============================================================
-- 4. ADD invited_by (NULLABLE for self-created memberships)
-- ============================================================
ALTER TABLE team_members 
ADD COLUMN invited_by UUID REFERENCES users(id);

-- ============================================================
-- 5. ADD UNIQUE CONSTRAINT
-- ============================================================
ALTER TABLE team_members 
ADD CONSTRAINT unique_member_institution 
UNIQUE (member_id, institution_id);

-- ============================================================
-- 6. DROP OLD INDEXES
-- ============================================================
DROP INDEX IF EXISTS idx_team_members_account_id;
DROP INDEX IF EXISTS idx_team_members_member_id;
DROP INDEX IF EXISTS idx_team_members_role;
DROP INDEX IF EXISTS idx_team_members_slug;
DROP INDEX IF EXISTS idx_team_members_user_id;

-- ============================================================
-- 7. CREATE NEW INDEXES
-- ============================================================
CREATE INDEX idx_team_members_member_id ON team_members(member_id);
CREATE INDEX idx_team_members_institution_id ON team_members(institution_id);
CREATE INDEX idx_team_members_invited_by ON team_members(invited_by);
CREATE INDEX idx_team_members_is_active ON team_members(is_active);
CREATE INDEX idx_team_members_created_by ON team_members(created_by);

-- ============================================================
-- 8. UPDATE COMMENTS
-- ============================================================
COMMENT ON TABLE team_members IS 'Links users to teams (institutions or personal)';
COMMENT ON COLUMN team_members.id IS 'Unique membership identifier';
COMMENT ON COLUMN team_members.member_id IS 'The user who is a member';
COMMENT ON COLUMN team_members.institution_id IS 'The institution this member belongs to (NULL for personal teams)';
COMMENT ON COLUMN team_members.invited_by IS 'The user who invited this member (NULL if self-created)';
COMMENT ON COLUMN team_members.is_active IS 'Whether the membership is active';
COMMENT ON COLUMN team_members.joined_at IS 'When the member joined the team';
COMMENT ON COLUMN team_members.created_by IS 'User who created this membership record';
COMMENT ON COLUMN team_members.created_at IS 'Record creation timestamp';
COMMENT ON COLUMN team_members.updated_at IS 'Record last update timestamp';
COMMENT ON COLUMN team_members.deleted_at IS 'Soft delete timestamp (NULL if not deleted)';

-- ============================================================
-- 9. UPDATE EXISTING RECORDS
-- ============================================================
-- For existing institution memberships, keep institution_id as is
-- No need to update anything

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- ============================================================
-- DOWN: Revert team_members changes
-- ============================================================

-- Drop unique constraint
ALTER TABLE team_members DROP CONSTRAINT IF EXISTS unique_member_institution;

-- Drop new indexes
DROP INDEX IF EXISTS idx_team_members_member_id;
DROP INDEX IF EXISTS idx_team_members_institution_id;
DROP INDEX IF EXISTS idx_team_members_invited_by;
DROP INDEX IF EXISTS idx_team_members_is_active;
DROP INDEX IF EXISTS idx_team_members_created_by;

-- Remove new columns
ALTER TABLE team_members 
DROP COLUMN IF EXISTS institution_id,
DROP COLUMN IF EXISTS invited_by;

-- Rename member_id back to user_id
ALTER TABLE team_members RENAME COLUMN member_id TO user_id;

-- Restore old columns
ALTER TABLE team_members 
ADD COLUMN slug VARCHAR(50),
ADD COLUMN name VARCHAR(100),
ADD COLUMN display_name VARCHAR(150),
ADD COLUMN member_id UUID REFERENCES users(id),
ADD COLUMN role VARCHAR(50) DEFAULT 'team_member',
ADD COLUMN job_title VARCHAR(100);

-- Restore old indexes
CREATE INDEX idx_team_members_account_id ON team_members(user_id);
CREATE INDEX idx_team_members_member_id ON team_members(member_id);
CREATE INDEX idx_team_members_role ON team_members(role);
CREATE INDEX idx_team_members_slug ON team_members(slug);

-- +goose StatementEnd