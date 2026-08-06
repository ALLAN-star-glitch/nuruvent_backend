-- +goose Up
-- +goose StatementBegin

-- Create attendance_statuses table
CREATE TABLE IF NOT EXISTS attendance_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(20),
    icon VARCHAR(50),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    can_issue_certificate BOOLEAN DEFAULT false,
    is_final BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_attendance_statuses_name ON attendance_statuses (name);
CREATE INDEX idx_attendance_statuses_is_active ON attendance_statuses (is_active);
CREATE INDEX idx_attendance_statuses_sort_order ON attendance_statuses (sort_order);

-- Create update trigger for updated_at
CREATE OR REPLACE FUNCTION update_attendance_statuses_updated_at() 
RETURNS TRIGGER LANGUAGE plpgsql AS '
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
';

DROP TRIGGER IF EXISTS update_attendance_statuses_updated_at ON attendance_statuses;
CREATE TRIGGER update_attendance_statuses_updated_at
    BEFORE UPDATE ON attendance_statuses
    FOR EACH ROW
    EXECUTE FUNCTION update_attendance_statuses_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop trigger
DROP TRIGGER IF EXISTS update_attendance_statuses_updated_at ON attendance_statuses;
DROP FUNCTION IF EXISTS update_attendance_statuses_updated_at();

-- Drop table
DROP TABLE IF EXISTS attendance_statuses;

-- +goose StatementEnd