// internal/modules/registration/infrastructure/postgres/waitlist_status_model.go

package postgres

import "time"

type WaitlistStatusModel struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
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

func (WaitlistStatusModel) TableName() string {
	return "waitlist_statuses"
}