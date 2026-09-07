-- +goose Up
-- +goose StatementBegin
-- Replace institution_id with team_id in events table

-- 1. Add team_id column (allow NULL since there's no data)
ALTER TABLE events ADD COLUMN team_id UUID;

-- 2. Create index on team_id
CREATE INDEX idx_events_team_id ON events(team_id);

-- 3. Add foreign key constraint to teams table
ALTER TABLE events ADD CONSTRAINT fk_events_team_id 
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL;

-- 4. Drop the old institution_id column
ALTER TABLE events DROP COLUMN institution_id;

-- 5. Drop the old indexes on institution_id
DROP INDEX IF EXISTS idx_events_institution_id;
DROP INDEX IF EXISTS idx_events_institution_id_new;

-- 6. We're NOT enforcing NOT NULL since there's no data
-- The team_id will remain NULL until events are created with proper teams
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Rollback: Restore institution_id and remove team_id

-- 1. Add back institution_id column
ALTER TABLE events ADD COLUMN institution_id UUID;

-- 2. Add index back
CREATE INDEX idx_events_institution_id ON events(institution_id);

-- 3. Add foreign key constraint back to institutions table
ALTER TABLE events ADD CONSTRAINT fk_events_institution_id 
    FOREIGN KEY (institution_id) REFERENCES institutions(id) ON DELETE SET NULL;

-- 4. Drop team_id column
DROP INDEX IF EXISTS idx_events_team_id;
ALTER TABLE events DROP COLUMN team_id;
-- +goose StatementEnd