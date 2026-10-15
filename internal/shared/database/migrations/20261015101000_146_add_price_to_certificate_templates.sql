-- +goose Up

-- Template-level certificate price override.
-- NULL means "no override" — fall back to the account default.
-- 0 means "explicitly free".
ALTER TABLE certificate_templates
  ADD COLUMN price_minor BIGINT NULL;


-- +goose Down

ALTER TABLE certificate_templates
  DROP COLUMN IF EXISTS price_minor;