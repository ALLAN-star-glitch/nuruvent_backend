// internal/modules/certificate/domain/tier_features.go

package domain

// ============================================================
// TIER FEATURE MATRIX
// ============================================================
//
// System templates are the pricing tiers Nuruvent offers:
//
//   default      → Basic         (KES 20)
//   professional → Professional  (KES 40)
//   premium      → Premium       (KES 60)
//
// Each tier permits a different set of design fields. Fields not
// listed for a tier are stripped on save — the UI hides them, but the
// service enforces the rule so no API call can bypass it.
//
// An empty field map means "everything is allowed" (used by Premium).

// TierAllowedFields maps a tier's system slug to the design fields it
// permits.
var TierAllowedFields = map[SystemTemplateSlug]map[string]bool{
	SystemTemplateDefault: {
		"background_color": true,
		"text_color":       true,
		"accent_color":     true,
		"font_family":      true,
		"title_text":       true,
		"subtitle_text":    true,
		"body_text":        true,
		"footer_text":      true,
		"qr_position":      true,
		"border_style":     true, // single only — see TierAllowedBorderStyles
		"border_color":     true,
	},
	SystemTemplateProfessional: {
		"background_color": true,
		"text_color":       true,
		"accent_color":     true,
		"font_family":      true,
		"title_text":       true,
		"subtitle_text":    true,
		"body_text":        true,
		"footer_text":      true,
		"qr_position":      true,
		"border_style":     true,
		"border_color":     true,
		"logo_url":         true,
		"signature_url":    true,
		"signatory_name":   true,
		"signatory_title":  true,
		"signature_line":   true,
		"verify_badge":     true,
	},
	SystemTemplatePremium: {
		// empty map → everything allowed
	},
}

// TierAllowedBorderStyles lists the border styles each tier permits.
// An empty string is the "no border" sentinel.
var TierAllowedBorderStyles = map[SystemTemplateSlug][]string{
	SystemTemplateDefault:      {"", "none", "single"},
	SystemTemplateProfessional: {"", "none", "single", "double"},
	SystemTemplatePremium:      {"", "none", "single", "double"},
}

// IsFieldAllowed reports whether the given tier permits a design field.
//
// Unknown tiers are treated as unrestricted so account templates created
// before the tier system, or without a tier, keep their existing designs.
func IsFieldAllowed(tier SystemTemplateSlug, field string) bool {
	if tier == "" {
		return true
	}
	allowed, ok := TierAllowedFields[tier]
	if !ok {
		return true
	}
	if len(allowed) == 0 {
		return true
	}
	return allowed[field]
}

// IsBorderStyleAllowed reports whether the tier permits a border style.
func IsBorderStyleAllowed(tier SystemTemplateSlug, style string) bool {
	if tier == "" {
		return true
	}
	styles, ok := TierAllowedBorderStyles[tier]
	if !ok {
		return true
	}
	for _, s := range styles {
		if s == style {
			return true
		}
	}
	return false
}

// SanitizeDesign strips design fields the tier doesn't permit.
//
// Called by CreateTemplate and UpdateTemplate before persisting, so
// no caller can smuggle a logo into a Basic template or a double
// border into a Basic one.
func SanitizeDesign(tier SystemTemplateSlug, d TemplateDesign) TemplateDesign {
	if !IsFieldAllowed(tier, "logo_url") {
		d.LogoURL = ""
	}
	if !IsFieldAllowed(tier, "signature_url") {
		d.SignatureURL = ""
	}
	if !IsFieldAllowed(tier, "signatory_name") {
		d.SignatoryName = ""
	}
	if !IsFieldAllowed(tier, "signatory_title") {
		d.SignatoryTitle = ""
	}
	if !IsFieldAllowed(tier, "signature_line") {
		d.SignatureLine = false
	}
	if !IsFieldAllowed(tier, "verify_badge") {
		d.VerifyBadge = false
	}
	if !IsFieldAllowed(tier, "seal_url") {
		d.SealURL = ""
	}
	if !IsFieldAllowed(tier, "background_url") {
		d.BackgroundURL = ""
	}
	if !IsFieldAllowed(tier, "watermark_text") {
		d.WatermarkText = ""
	}
	if !IsBorderStyleAllowed(tier, d.BorderStyle) {
		d.BorderStyle = ""
	}
	return d
}