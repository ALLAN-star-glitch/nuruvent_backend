-- +goose Up

-- ============================================================
-- Seed tier prices for system certificate templates.
--
-- Basic      KES 20.00  → 2000 minor
-- Standard   KES 40.00  → 4000 minor
-- Premium    KES 60.00  → 6000 minor
-- ============================================================

UPDATE certificate_templates
SET price_minor = 2000
WHERE slug = 'default'
  AND is_system = TRUE
  AND price_minor IS NULL;

UPDATE certificate_templates
SET price_minor = 4000
WHERE slug = 'professional'
  AND is_system = TRUE
  AND price_minor IS NULL;

UPDATE certificate_templates
SET price_minor = 6000
WHERE slug = 'premium'
  AND is_system = TRUE
  AND price_minor IS NULL;


-- +goose Down

UPDATE certificate_templates
SET price_minor = NULL
WHERE slug IN ('default', 'professional', 'premium')
  AND is_system = TRUE
  AND price_minor IN (2000, 4000, 6000);