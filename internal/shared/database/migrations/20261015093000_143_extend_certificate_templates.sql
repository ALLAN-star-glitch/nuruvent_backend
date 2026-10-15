-- +goose Up

-- Add account-scoped columns for user templates.
ALTER TABLE certificate_templates
  ADD COLUMN account_id UUID NULL REFERENCES accounts(id) ON DELETE RESTRICT,
  ADD COLUMN design     JSONB NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN is_default BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN is_system  BOOLEAN NOT NULL DEFAULT FALSE;

-- Existing rows (seeded templates) are system templates.
UPDATE certificate_templates SET is_system = TRUE WHERE is_system = FALSE;

-- Slug is only required for system templates going forward.
ALTER TABLE certificate_templates
  ALTER COLUMN slug DROP NOT NULL;

-- Drop the plain unique on slug; replace with partial (only system rows).
ALTER TABLE certificate_templates
  DROP CONSTRAINT certificate_templates_slug_key;

CREATE UNIQUE INDEX uniq_certificate_templates_system_slug
  ON certificate_templates (slug)
  WHERE is_system = TRUE;

-- One default per account (partial — only account rows).
CREATE UNIQUE INDEX uniq_certificate_templates_account_default
  ON certificate_templates (account_id)
  WHERE is_system = FALSE AND is_default = TRUE;

-- Account listing index.
CREATE INDEX idx_certificate_templates_account
  ON certificate_templates (account_id)
  WHERE is_system = FALSE;


-- +goose Down

DROP INDEX IF EXISTS idx_certificate_templates_account;
DROP INDEX IF EXISTS uniq_certificate_templates_account_default;
DROP INDEX IF EXISTS uniq_certificate_templates_system_slug;

ALTER TABLE certificate_templates
  ADD CONSTRAINT certificate_templates_slug_key UNIQUE (slug);

ALTER TABLE certificate_templates
  ALTER COLUMN slug SET NOT NULL;

ALTER TABLE certificate_templates
  DROP COLUMN is_system,
  DROP COLUMN is_default,
  DROP COLUMN design,
  DROP COLUMN account_id;