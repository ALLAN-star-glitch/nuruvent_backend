// internal/modules/certificate/domain/template.go

package domain

import (
	"strings"
	"time"
)

// SystemTemplateSlug identifies templates Nuruvent ships with. They are
// available to every account that has not defined its own default.
type SystemTemplateSlug string

const (
	SystemTemplateDefault      SystemTemplateSlug = "default"
	SystemTemplateProfessional SystemTemplateSlug = "professional"
	SystemTemplatePremium      SystemTemplateSlug = "premium"
)

// AllSystemTemplates lists every system template slug.
var AllSystemTemplates = []SystemTemplateSlug{
	SystemTemplateDefault,
	SystemTemplateProfessional,
	SystemTemplatePremium,
}

// IsValid reports whether the slug is a known system template.
func (s SystemTemplateSlug) IsValid() bool {
	for _, t := range AllSystemTemplates {
		if t == s {
			return true
		}
	}
	return false
}

// PriceForSlug returns the canonical price (in minor units) for a
// system template slug. This is the single source of truth for tier
// pricing — system templates derive their price from their slug, never
// the other way around.
func PriceForSlug(s SystemTemplateSlug) int64 {
	switch s {
	case SystemTemplateDefault:
		return 2000
	case SystemTemplateProfessional:
		return 4000
	case SystemTemplatePremium:
		return 6000
	}
	return 0
}

// DefaultBackgroundOpacity is applied when the caller does not set a
// background opacity (0 / unset). Chosen to keep text legible on top of
// full-bleed patterns.
const DefaultBackgroundOpacity = 8

// TemplateDesign is the renderable design of a certificate template.
type TemplateDesign struct {
	// Visual
	BackgroundColor string
	TextColor       string
	AccentColor     string
	FontFamily      string

	// Images (object-storage references)
	LogoURL       string
	SealURL       string
	SignatureURL  string
	BackgroundURL string

	// Background image behavior
	BackgroundOpacity  int    // 0–100. 0 = use default (8)
	BackgroundSize     string // "cover" | "contain" | "auto"
	BackgroundPosition string // "center" | "top" | "bottom"

	// Signature printed name
	SignatoryName  string
	SignatoryTitle string

	// Layout slots
	TitleText    string
	SubtitleText string
	BodyText     string
	FooterText   string

	// QR placement hint
	QRPosition string

	// Border
	BorderStyle string
	BorderColor string

	// Signature
	SignatureLine bool

	// Watermark
	WatermarkText string

	// Badge
	VerifyBadge bool
}

// NormalizeBackground applies safe defaults and clamps to valid ranges.
// Call this before persisting or rendering. Safe to call on legacy
// designs that predate the background fields.
//
// Rules:
//   - Opacity <= 0 → DefaultBackgroundOpacity
//   - Opacity > 100 → 100
//   - Size → one of "cover", "contain", "auto" (default "cover")
//   - Position → one of "center", "top", "bottom" (default "center")
//   - BackgroundURL is cleared unless it starts with http:// or https://
func (d *TemplateDesign) NormalizeBackground() {
	if d.BackgroundOpacity <= 0 {
		d.BackgroundOpacity = DefaultBackgroundOpacity
	}
	if d.BackgroundOpacity > 100 {
		d.BackgroundOpacity = 100
	}

	d.BackgroundSize = normalizeEnum(d.BackgroundSize, "cover", "contain", "auto")
	d.BackgroundPosition = normalizeEnum(d.BackgroundPosition, "center", "top", "bottom")

	if !isHTTPURL(d.BackgroundURL) {
		d.BackgroundURL = ""
	}
}

// normalizeEnum returns v if it matches one of allowed (case-insensitive),
// otherwise returns allowed[0].
func normalizeEnum(v string, allowed ...string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	for _, a := range allowed {
		if v == a {
			return v
		}
	}
	return allowed[0]
}

// isHTTPURL reports whether s is a plausible http(s) URL.
func isHTTPURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// Template is a reusable certificate design owned by an account.
type Template struct {
	ID string

	// AccountID is nil for system templates, and set for account-owned templates.
	AccountID *string

	Name        string
	Description string
	Design      TemplateDesign

	// PriceMinor is the optional template-level certificate price.
	// For system templates this is always derived from SystemSlug via
	// PriceForSlug — never set directly.
	PriceMinor *int64

	// SystemSlug is non-empty for templates Nuruvent ships with. Such
	// templates are shared and not owned by any single account.
	SystemSlug SystemTemplateSlug

	// IsDefault marks the account's default template for new issuances.
	IsDefault bool

	// IsSystem mirrors SystemSlug != "" but is stored so listings and
	// queries can filter without parsing the slug.
	IsSystem bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Validate checks the template's invariants.
//
// System templates MUST have a valid slug — the slug defines the tier,
// and the tier defines what design features are permitted. Account
// templates must NOT carry a slug.
func (t Template) Validate() error {
	if t.IsSystem {
		// System templates: slug is mandatory and must be one of the
		// three known tiers.
		if t.SystemSlug == "" {
			return ErrSystemTemplateSlugRequired
		}
		if !t.SystemSlug.IsValid() {
			return ErrInvalidSystemTemplateSlug
		}
	} else {
		// Account templates: no slug, must have an owning account.
		if t.SystemSlug != "" {
			return ErrInvalidSystemTemplateSlug
		}
		if t.AccountID == nil || strings.TrimSpace(*t.AccountID) == "" {
			return ErrAccountIDRequired
		}
	}

	if strings.TrimSpace(t.Name) == "" {
		return ErrTemplateNameRequired
	}
	return nil
}

// IsDeleted reports whether the template is soft-deleted.
func (t Template) IsDeleted() bool { return t.DeletedAt != nil }