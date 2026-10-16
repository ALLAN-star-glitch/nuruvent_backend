-- +goose Up

UPDATE certificate_templates
SET design = jsonb_build_object(
    'background_color', COALESCE(design->>'BackgroundColor', design->>'background_color', '#FFFFFF'),
    'text_color',       COALESCE(design->>'TextColor',       design->>'text_color',       '#1A1A1A'),
    'accent_color',     COALESCE(design->>'AccentColor',     design->>'accent_color',     '#0A6EBD'),
    'font_family',      COALESCE(design->>'FontFamily',      design->>'font_family',      'Helvetica'),
    'logo_url',         COALESCE(design->>'LogoURL',         design->>'logo_url',         ''),
    'seal_url',         COALESCE(design->>'SealURL',         design->>'seal_url',         ''),
    'signature_url',    COALESCE(design->>'SignatureURL',    design->>'signature_url',    ''),
    'background_url',   COALESCE(design->>'BackgroundURL',   design->>'background_url',   ''),
    'signatory_name',   COALESCE(design->>'SignatoryName',   design->>'signatory_name',   ''),
    'signatory_title',  COALESCE(design->>'SignatoryTitle',  design->>'signatory_title',  ''),
    'title_text',       COALESCE(design->>'TitleText',       design->>'title_text',       'Certificate of Completion'),
    'subtitle_text',    COALESCE(design->>'SubtitleText',    design->>'subtitle_text',    'This certifies that'),
    'body_text',        COALESCE(design->>'BodyText',        design->>'body_text',        'has successfully completed {{context_title}} on {{context_date}}.'),
    'footer_text',      COALESCE(design->>'FooterText',      design->>'footer_text',      'Verify at nuruvent.com/verify'),
    'qr_position',      COALESCE(design->>'QRPosition',      design->>'qr_position',      'bottom-right')
)
WHERE design IS NOT NULL AND design != '{}'::jsonb;


-- +goose Down

SELECT 1;