-- +goose Up
-- +goose StatementBegin

-- ================================================
-- 1. Create institution_types table
-- ================================================
CREATE TABLE IF NOT EXISTS institution_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(20),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    meta_title VARCHAR(150),
    meta_description TEXT,
    slug VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_institution_types_name ON institution_types (name);
CREATE INDEX idx_institution_types_slug ON institution_types (slug);
CREATE INDEX idx_institution_types_is_active ON institution_types (is_active);
CREATE INDEX idx_institution_types_sort_order ON institution_types (sort_order);

-- ================================================
-- 2. Create accounts table
-- ================================================
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    account_type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    professional_type VARCHAR(50),
    institution_id UUID,
    email_verified BOOLEAN DEFAULT false,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    identity_verified BOOLEAN DEFAULT false,
    identity_verified_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_accounts_email ON accounts (email);
CREATE INDEX idx_accounts_account_type ON accounts (account_type);
CREATE INDEX idx_accounts_institution_id ON accounts (institution_id);
CREATE INDEX idx_accounts_email_verified ON accounts (email_verified);
CREATE INDEX idx_accounts_identity_verified ON accounts (identity_verified);
CREATE INDEX idx_accounts_is_active ON accounts (is_active);

-- ================================================
-- 3. Create institutions table
-- ================================================
CREATE TABLE IF NOT EXISTS institutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(50),
    institution_type_id UUID NOT NULL,
    description TEXT,
    logo VARCHAR(500),
    website VARCHAR(255),
    address TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_institutions_name ON institutions (name);
CREATE INDEX idx_institutions_email ON institutions (email);
CREATE INDEX idx_institutions_type_id ON institutions (institution_type_id);
CREATE INDEX idx_institutions_is_active ON institutions (is_active);

-- ================================================
-- 4. Create institution_admins table
-- ================================================
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

-- Create indexes
CREATE INDEX idx_institution_admins_institution_id ON institution_admins (institution_id);
CREATE INDEX idx_institution_admins_account_id ON institution_admins (account_id);
CREATE INDEX idx_institution_admins_is_active ON institution_admins (is_active);

-- ================================================
-- 5. Add foreign key constraints
-- ================================================
ALTER TABLE institutions ADD CONSTRAINT fk_institutions_type 
    FOREIGN KEY (institution_type_id) REFERENCES institution_types(id) ON DELETE RESTRICT;

ALTER TABLE accounts ADD CONSTRAINT fk_accounts_institution 
    FOREIGN KEY (institution_id) REFERENCES institutions(id) ON DELETE SET NULL;

ALTER TABLE institution_admins ADD CONSTRAINT fk_institution_admins_institution 
    FOREIGN KEY (institution_id) REFERENCES institutions(id) ON DELETE CASCADE;

ALTER TABLE institution_admins ADD CONSTRAINT fk_institution_admins_account 
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE;

-- ================================================
-- 6. Create update triggers for updated_at
-- ================================================
CREATE OR REPLACE FUNCTION update_institution_types_updated_at() 
RETURNS TRIGGER LANGUAGE plpgsql AS '
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
';

CREATE OR REPLACE FUNCTION update_accounts_updated_at() 
RETURNS TRIGGER LANGUAGE plpgsql AS '
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
';

CREATE OR REPLACE FUNCTION update_institutions_updated_at() 
RETURNS TRIGGER LANGUAGE plpgsql AS '
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
';

CREATE OR REPLACE FUNCTION update_institution_admins_updated_at() 
RETURNS TRIGGER LANGUAGE plpgsql AS '
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
';

-- Drop existing triggers if they exist
DROP TRIGGER IF EXISTS update_institution_types_updated_at ON institution_types;
DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;
DROP TRIGGER IF EXISTS update_institutions_updated_at ON institutions;
DROP TRIGGER IF EXISTS update_institution_admins_updated_at ON institution_admins;

-- Create triggers
CREATE TRIGGER update_institution_types_updated_at
    BEFORE UPDATE ON institution_types
    FOR EACH ROW
    EXECUTE FUNCTION update_institution_types_updated_at();

CREATE TRIGGER update_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_accounts_updated_at();

CREATE TRIGGER update_institutions_updated_at
    BEFORE UPDATE ON institutions
    FOR EACH ROW
    EXECUTE FUNCTION update_institutions_updated_at();

CREATE TRIGGER update_institution_admins_updated_at
    BEFORE UPDATE ON institution_admins
    FOR EACH ROW
    EXECUTE FUNCTION update_institution_admins_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- ================================================
-- Drop triggers
-- ================================================
DROP TRIGGER IF EXISTS update_institution_types_updated_at ON institution_types;
DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;
DROP TRIGGER IF EXISTS update_institutions_updated_at ON institutions;
DROP TRIGGER IF EXISTS update_institution_admins_updated_at ON institution_admins;

DROP FUNCTION IF EXISTS update_institution_types_updated_at();
DROP FUNCTION IF EXISTS update_accounts_updated_at();
DROP FUNCTION IF EXISTS update_institutions_updated_at();
DROP FUNCTION IF EXISTS update_institution_admins_updated_at();

-- ================================================
-- Drop foreign key constraints
-- ================================================
ALTER TABLE institutions DROP CONSTRAINT IF EXISTS fk_institutions_type;
ALTER TABLE accounts DROP CONSTRAINT IF EXISTS fk_accounts_institution;
ALTER TABLE institution_admins DROP CONSTRAINT IF EXISTS fk_institution_admins_institution;
ALTER TABLE institution_admins DROP CONSTRAINT IF EXISTS fk_institution_admins_account;

-- ================================================
-- Drop tables
-- ================================================
DROP TABLE IF EXISTS institution_admins;
DROP TABLE IF EXISTS institutions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS institution_types;

-- ================================================
-- Drop indexes (optional, as they're dropped with tables)
-- ================================================
DROP INDEX IF EXISTS idx_institution_types_name;
DROP INDEX IF EXISTS idx_institution_types_slug;
DROP INDEX IF EXISTS idx_institution_types_is_active;
DROP INDEX IF EXISTS idx_institution_types_sort_order;

DROP INDEX IF EXISTS idx_accounts_email;
DROP INDEX IF EXISTS idx_accounts_account_type;
DROP INDEX IF EXISTS idx_accounts_institution_id;
DROP INDEX IF EXISTS idx_accounts_email_verified;
DROP INDEX IF EXISTS idx_accounts_identity_verified;
DROP INDEX IF EXISTS idx_accounts_is_active;

DROP INDEX IF EXISTS idx_institutions_name;
DROP INDEX IF EXISTS idx_institutions_email;
DROP INDEX IF EXISTS idx_institutions_type_id;
DROP INDEX IF EXISTS idx_institutions_is_active;

DROP INDEX IF EXISTS idx_institution_admins_institution_id;
DROP INDEX IF EXISTS idx_institution_admins_account_id;
DROP INDEX IF EXISTS idx_institution_admins_is_active;

-- +goose StatementEnd