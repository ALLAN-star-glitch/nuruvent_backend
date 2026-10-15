// internal/modules/certificate/infrastructure/postgres/delivery_model.go

package postgres

import "time"

// DeliveryModel is the GORM representation of the certificate_deliveries
// table.
type DeliveryModel struct {
	ID            string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CertificateID string `gorm:"type:uuid;index;not null"`
	Channel       string `gorm:"type:varchar(20);not null"`
	Recipient     string `gorm:"type:varchar(255);not null"`
	Status        string `gorm:"type:varchar(20);not null;default:'pending'"`
	Error         *string `gorm:"type:text"`
	Attempts      int     `gorm:"not null;default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName pins the model to the table created by migration 142.
func (DeliveryModel) TableName() string { return "certificate_deliveries" }