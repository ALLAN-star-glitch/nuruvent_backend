// internal/modules/certificate/infrastructure/postgres/template_model.go

package postgres

import "time"

// TemplateModel is the GORM representation of the certificate_templates
// table.
//
// The table is shared with the events module (which references it via
// events.certificate_template_id) and holds both:
//
//   - system templates (is_system = TRUE, slug set, system_slug set, account_id NULL)
//   - account templates (is_system = FALSE, account_id set, slug NULL, system_slug '')
type TemplateModel struct {
	ID          string  `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Slug        *string `gorm:"type:varchar(100);index"`
	Name        string  `gorm:"type:varchar(100);not null"`
	DisplayName *string `gorm:"type:varchar(100)"`
	Description *string `gorm:"type:text"`
	PreviewURL  *string `gorm:"type:varchar(500)"`
	IsActive    bool    `gorm:"default:true"`

	AccountID *string `gorm:"type:uuid;index"`
	Design    JSONB   `gorm:"type:jsonb;not null;default:'{}'"`
	IsDefault bool    `gorm:"not null;default:false"`
	IsSystem  bool    `gorm:"not null;default:false"`

	// SystemSlug is the tier identifier for system templates:
	// "default", "professional", or "premium". Empty for account
	// templates. This is distinct from Slug — Slug is a URL-safe
	// identifier used for stable lookup, SystemSlug is the tier.
	SystemSlug string `gorm:"column:system_slug;type:varchar(20);not null;default:''"`

	PriceMinor *int64 `gorm:"column:price_minor"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`
}

func (TemplateModel) TableName() string { return "certificate_templates" }