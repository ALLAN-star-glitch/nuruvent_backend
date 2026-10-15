// internal/modules/certificate/infrastructure/postgres/certificate_type_model.go

package postgres

import "time"

// CertificateTypeModel is the GORM representation of the
// certificate_types lookup table.
type CertificateTypeModel struct {
	ID          string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Slug        string `gorm:"uniqueIndex;type:varchar(100);not null"`
	Name        string `gorm:"type:varchar(100);not null"`
	DisplayName string `gorm:"type:varchar(100)"`
	IsActive    bool   `gorm:"default:true"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`
}

func (CertificateTypeModel) TableName() string { return "certificate_types" }