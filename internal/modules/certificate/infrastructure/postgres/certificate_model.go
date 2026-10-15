// internal/modules/certificate/infrastructure/postgres/certificate_model.go

package postgres

import "time"

// CertificateModel is the GORM representation of the certificates table.
type CertificateModel struct {
	ID               string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AccountID        string `gorm:"type:uuid;index;not null"`
	TemplateID       string `gorm:"type:uuid;index;not null"`
	VerificationCode string `gorm:"type:varchar(64);uniqueIndex;not null"`

	// Recipient snapshot
	RecipientUserID *string `gorm:"type:uuid;index"`
	RecipientName   string  `gorm:"type:varchar(255);not null"`
	RecipientEmail  string  `gorm:"type:varchar(255);not null"`

	// Opaque context reference
	ContextType string `gorm:"type:varchar(50);index;not null"`
	ContextID   string `gorm:"type:uuid;index;not null"`

	// Free-form snapshot
	Metadata JSONB `gorm:"type:jsonb;not null;default:'{}'"`

	// Timeline
	CompletedAt time.Time `gorm:"not null"`
	IssuedAt    time.Time `gorm:"not null"`

	// Payment
	PaymentReference *string `gorm:"type:varchar(255)"`

	// Lifecycle
	Status           string     `gorm:"type:varchar(20);not null;default:'active'"`
	RevokedAt        *time.Time
	RevocationReason *string `gorm:"type:text"`

	// Cached PDF
	PDFURL         *string `gorm:"type:varchar(500)"`
	PDFGeneratedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName pins the model to the table created by migration 141.
func (CertificateModel) TableName() string { return "certificates" }