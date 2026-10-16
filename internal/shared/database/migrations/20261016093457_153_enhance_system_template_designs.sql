-- +goose Up

-- ============================================================
-- Enhance system certificate template designs.
--
-- Adds structural fields (border, signature line, verify badge,
-- watermark) that the PDF renderer now supports. Each tier gets a
-- different level of ornamentation, making them visually distinct
-- and justifying the price difference.
--
-- Basic        — thin single border, no signature line, no badge
-- Professional — double border, signature line, verified badge
-- Premium      — double gold border, signature line, badge, watermark
-- ============================================================

UPDATE certificate_templates
SET design = design
  || '{"border_style":"single","border_color":"#2563EB","signature_line":false,"verify_badge":false}'::jsonb
WHERE slug = 'default' AND is_system = TRUE;

UPDATE certificate_templates
SET design = design
  || '{"border_style":"double","border_color":"#1E3A8A","signature_line":true,"verify_badge":true}'::jsonb
WHERE slug = 'professional' AND is_system = TRUE;

UPDATE certificate_templates
SET design = design
  || '{"border_style":"double","border_color":"#B45309","signature_line":true,"verify_badge":true,"watermark_text":"Official Certificate"}'::jsonb
WHERE slug = 'premium' AND is_system = TRUE;


-- +goose Down

UPDATE certificate_templates
SET design = design - 'border_style' - 'border_color' - 'signature_line' - 'verify_badge' - 'watermark_text'
WHERE is_system = TRUE;