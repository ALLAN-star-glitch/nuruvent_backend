// internal/modules/certificate/infrastructure/postgres/certificate_template_model.go

package postgres

import "time"

// CertificateTemplateModel is the GORM representation of the
// certificate_templates table.
//
// NOTE: this model is distinct from TemplateModel in template_model.go.
// TemplateModel is the account-template shape used by the
// TemplateRepository; CertificateTemplateModel is the seed-data shape
// used by the seeder to write system templates.
//
// They share the same physical table — the difference is which fields
// the caller populates. In practice, a single model would do; the two
// exist here only because the seeder was migrated separately and its
// column set is narrower.
//
// TableName pins the model to certificate_templates.
type CertificateTemplateModel struct {
	ID          string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Slug        string `gorm:"uniqueIndex;type:varchar(100);not null"`
	Name        string `gorm:"type:varchar(100);not null"`
	DisplayName string `gorm:"type:varchar(100)"`
	Description string `gorm:"type:text"`
	PreviewURL  string `gorm:"type:varchar(500)"`
	IsActive    bool   `gorm:"default:true"`

	// System-template flags added by migration 143.
	IsSystem bool `gorm:"not null;default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`
}

func (CertificateTemplateModel) TableName() string { return "certificate_templates" }