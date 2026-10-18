-- +goose Up
-- +goose StatementBegin

-- 1. Add the column. Default '' so existing rows decode as "no slug".
ALTER TABLE certificate_templates
  ADD COLUMN system_slug VARCHAR(20) NOT NULL DEFAULT '';

-- 2. Backfill existing system templates by price.
--    2000 → default, 4000 → professional, 6000 → premium.
UPDATE certificate_templates
SET system_slug = CASE price_minor
    WHEN 2000 THEN 'default'
    WHEN 4000 THEN 'professional'
    WHEN 6000 THEN 'premium'
END
WHERE is_system = true
  AND price_minor IN (2000, 4000, 6000);

-- 3. Enforce allowed values at the DB level.
ALTER TABLE certificate_templates
  ADD CONSTRAINT chk_system_slug_valid
  CHECK (
    system_slug = ''
    OR system_slug IN ('default', 'professional', 'premium')
  );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE certificate_templates DROP CONSTRAINT IF EXISTS chk_system_slug_valid;
ALTER TABLE certificate_templates DROP COLUMN IF EXISTS system_slug;
-- +goose StatementEnd