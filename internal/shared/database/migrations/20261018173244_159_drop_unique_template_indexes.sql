-- +goose Up
-- +goose StatementBegin

-- Drop the unique indexes that prevent multiple system templates
-- per tier and multiple templates sharing a slug.
--
-- Why: the app allows multiple templates of the same tier (e.g. two
-- different "professional" designs). Tier uniqueness is not a schema
-- invariant — it's a UI choice. Keeping these indexes forces the
-- backend to translate "duplicate key" into the wrong error
-- ("certificate already exists") and blocks valid template creation.
DROP INDEX IF EXISTS idx_system_templates_unique_slug;
DROP INDEX IF EXISTS uniq_certificate_templates_system_slug;
DROP INDEX IF EXISTS idx_certificate_templates_slug;

-- Non-unique replacement so slug lookups stay fast if anything queries
-- by slug later.
CREATE INDEX IF NOT EXISTS idx_certificate_templates_slug
  ON certificate_templates (slug);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_certificate_templates_slug;
-- +goose StatementEnd