// internal/modules/registration/infrastructure/postgres/registration_status_model.go

package postgres

import "time"

type RegistrationStatusModel struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Slug        string    `gorm:"uniqueIndex;not null"`
	Name        string    `gorm:"not null"`
	DisplayName string    `gorm:"not null"`
	Description string
	Color       string
	Icon        string
	SortOrder   int       `gorm:"not null;default:0"`
	IsFinal     bool      `gorm:"not null;default:false"`
	IsActive    bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (RegistrationStatusModel) TableName() string {
	return "registration_statuses"
}