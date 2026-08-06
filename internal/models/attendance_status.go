// internal/models/attendance_status.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendanceStatus struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`          // URL-safe: "full-attendance"
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`                // Canonical: "Full Attendance"
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`        // UI: "✅ Full Attendance"
	Description string         `gorm:"type:text" json:"description"`
	Color       string         `gorm:"type:varchar(20)" json:"color"`
	Icon        string         `gorm:"type:varchar(50)" json:"icon"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	
	// Status properties
	CanIssueCertificate bool `gorm:"default:false" json:"can_issue_certificate"`
	IsFinal             bool `gorm:"default:false" json:"is_final"`
	
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

