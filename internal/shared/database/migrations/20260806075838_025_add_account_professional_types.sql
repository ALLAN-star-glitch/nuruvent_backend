-- +goose Up
-- +goose StatementBegin

-- Create account_types table
CREATE TABLE IF NOT EXISTS account_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(150),
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(20),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_account_types_slug ON account_types (slug);

-- Create professional_types table
CREATE TABLE IF NOT EXISTS professional_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(150),
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(20),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    can_host BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_professional_types_slug ON professional_types (slug);

-- Add foreign keys to accounts
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS account_type_id UUID;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS professional_type_id UUID;

-- Add foreign key constraints
ALTER TABLE accounts ADD CONSTRAINT fk_accounts_account_type 
    FOREIGN KEY (account_type_id) REFERENCES account_types(id) ON DELETE RESTRICT;

ALTER TABLE accounts ADD CONSTRAINT fk_accounts_professional_type 
    FOREIGN KEY (professional_type_id) REFERENCES professional_types(id) ON DELETE SET NULL;

CREATE INDEX idx_accounts_account_type_id ON accounts (account_type_id);
CREATE INDEX idx_accounts_professional_type_id ON accounts (professional_type_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE accounts DROP CONSTRAINT IF EXISTS fk_accounts_account_type;
ALTER TABLE accounts DROP CONSTRAINT IF EXISTS fk_accounts_professional_type;

DROP INDEX IF EXISTS idx_accounts_account_type_id;
DROP INDEX IF EXISTS idx_accounts_professional_type_id;

ALTER TABLE accounts DROP COLUMN IF EXISTS account_type_id;
ALTER TABLE accounts DROP COLUMN IF EXISTS professional_type_id;

DROP TABLE IF EXISTS professional_types;
DROP TABLE IF EXISTS account_types;

-- +goose StatementEnd