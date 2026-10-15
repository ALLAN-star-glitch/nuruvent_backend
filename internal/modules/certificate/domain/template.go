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

// TemplateDesign is the renderable design of a certificate template.
//
// The domain owns the shape of this struct; the GORM adapter is
// responsible for marshaling it to and from the JSONB "design" column.
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

	// Signature printed name
	SignatoryName  string
	SignatoryTitle string

	// Layout slots — placeholders in the rendered PDF. Each is a
	// template string that may reference Metadata keys, e.g.
	// "Awarded to {{recipient_name}}".
	TitleText    string
	SubtitleText string
	BodyText     string
	FooterText   string

	// QR placement hint (renderer-defined coordinates)
	QRPosition string
}

// Template is a reusable certificate design owned by an account.
type Template struct {
	ID          string
	AccountID   string
	Name        string
	Description string
	Design      TemplateDesign

	// PriceMinor is the optional template-level certificate price.
	//
	//   nil = no override; use the account's default price
	//   0   = explicitly free for certificates using this template
	//   >0  = template-specific price in minor units
	//
	// Stored in certificate_templates.price_minor (nullable).
	PriceMinor *int64

	// SystemSlug is non-empty for templates Nuruvent ships with. Such
	// templates are shared and not owned by any single account.
	SystemSlug SystemTemplateSlug

	// IsDefault marks the account's default template for new issuances.
	// At most one default per account (enforced by repository).
	IsDefault bool

	// IsSystem mirrors SystemSlug != "" but is stored so listings and
	// queries can filter without parsing the slug.
	IsSystem bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Validate checks the template's invariants.
func (t Template) Validate() error {
	if strings.TrimSpace(t.AccountID) == "" && !t.IsSystem {
		return ErrAccountIDRequired
	}
	if strings.TrimSpace(t.Name) == "" {
		return ErrTemplateNameRequired
	}
	if t.IsSystem && !t.SystemSlug.IsValid() {
		return ErrInvalidSystemTemplateSlug
	}
	return nil
}

// IsDeleted reports whether the template is soft-deleted.
func (t Template) IsDeleted() bool { return t.DeletedAt != nil }