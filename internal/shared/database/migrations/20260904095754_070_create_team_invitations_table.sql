-- internal/shared/database/migrations/XXX_create_team_invitations_table.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS team_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    institution_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    invited_by UUID NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Foreign key constraints
    CONSTRAINT fk_team_invitations_institution 
        FOREIGN KEY (institution_id) 
        REFERENCES institutions(id) ON DELETE CASCADE,
    CONSTRAINT fk_team_invitations_invited_by 
        FOREIGN KEY (invited_by) 
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_team_invitations_status 
        CHECK (status IN ('pending', 'accepted', 'declined', 'expired')),
    CONSTRAINT chk_team_invitations_role 
        CHECK (role IN ('account_admin', 'event_manager', 'team_member'))
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_team_invitations_email 
    ON team_invitations(email);
CREATE INDEX IF NOT EXISTS idx_team_invitations_institution_id 
    ON team_invitations(institution_id);
CREATE INDEX IF NOT EXISTS idx_team_invitations_token 
    ON team_invitations(token);
CREATE INDEX IF NOT EXISTS idx_team_invitations_status 
    ON team_invitations(status);
CREATE INDEX IF NOT EXISTS idx_team_invitations_expires_at 
    ON team_invitations(expires_at);
CREATE INDEX IF NOT EXISTS idx_team_invitations_deleted_at 
    ON team_invitations(deleted_at);
    
-- Composite index for common queries
CREATE INDEX IF NOT EXISTS idx_team_invitations_email_institution 
    ON team_invitations(email, institution_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS team_invitations CASCADE;
-- +goose StatementEnd