// internal/modules/registration/infrastructure/postgres/attendance_status_model.go

package postgres

import "time"

type AttendanceStatusModel struct {
	ID               string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Slug             string    `gorm:"uniqueIndex;not null"`
	Name             string    `gorm:"not null"`
	DisplayName      string    `gorm:"not null"`
	Description      string
	Color            string
	Icon             string
	SortOrder        int       `gorm:"not null;default:0"`
	IsFinal          bool      `gorm:"not null;default:false"`
	IsActive         bool      `gorm:"not null;default:true"`
	CountsAsAttended bool      `gorm:"not null;default:false"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (AttendanceStatusModel) TableName() string {
	return "attendance_statuses"
}