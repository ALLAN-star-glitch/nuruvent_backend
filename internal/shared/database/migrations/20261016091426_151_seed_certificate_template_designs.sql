-- +goose Up

-- ============================================================
-- Seed designs for system certificate templates.
--
-- The price seeding migration (149) sets only prices. Without
-- designs, all three system tiers render identically, defeating
-- the purpose of offering tiers.
--
-- Each tier gets a distinct colour scheme, font, and copy so the
-- difference is visible in previews and on issued certificates.
-- ============================================================

UPDATE certificate_templates
SET design = '{
  "background_color": "#FFFFFF",
  "text_color": "#1A1A1A",
  "accent_color": "#2563EB",
  "font_family": "Helvetica",
  "title_text": "Certificate of Completion",
  "subtitle_text": "This certifies that",
  "body_text": "has successfully completed {{context_title}} on {{context_date}}.",
  "footer_text": "Verify at nuruvent.com/verify",
  "qr_position": "bottom-right"
}'::jsonb
WHERE slug = 'default'
  AND is_system = TRUE;

UPDATE certificate_templates
SET design = '{
  "background_color": "#F8FAFC",
  "text_color": "#0F172A",
  "accent_color": "#1E3A8A",
  "font_family": "Times-Roman",
  "title_text": "Certificate of Professional Achievement",
  "subtitle_text": "This is to certify that",
  "body_text": "has successfully completed {{context_title}} on {{context_date}}.",
  "footer_text": "Verify this certificate at nuruvent.com/verify",
  "qr_position": "bottom-right"
}'::jsonb
WHERE slug = 'professional'
  AND is_system = TRUE;

UPDATE certificate_templates
SET design = '{
  "background_color": "#FFFBF0",
  "text_color": "#1A1A1A",
  "accent_color": "#B45309",
  "font_family": "Times-Roman",
  "title_text": "Certificate of Excellence",
  "subtitle_text": "Proudly presented to",
  "body_text": "for outstanding completion of {{context_title}} on {{context_date}}.",
  "footer_text": "Authenticity verifiable at nuruvent.com/verify",
  "qr_position": "bottom-right"
}'::jsonb
WHERE slug = 'premium'
  AND is_system = TRUE;


-- +goose Down

UPDATE certificate_templates
SET design = '{}'::jsonb
WHERE slug IN ('default', 'professional', 'premium')
  AND is_system = TRUE;