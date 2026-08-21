package postgres

import (
	"time"

	"gorm.io/gorm"
)

// ============================================================
// EVENT MODEL
// ============================================================

type EventModel struct {
	ID               string `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug             string `gorm:"uniqueIndex"`
	Name             string
	DisplayName      string
	Description      string
	EventTypeID      string `gorm:"index"`
	EventStatusID    string `gorm:"index"`
	ImageURL         string
	ThumbnailURL     string
	Date             time.Time
	Time             string
	Duration         int
	Price            float64
	CertificatePrice float64
	Location         string
	IsVirtual        bool
	ZoomLink         string
	MeetLink         string
	MaxAttendees     int
	CurrentAttendees int
	AccountID        string `gorm:"index"`
	CreatedBy        string `gorm:"index"`
	IsActive         bool   `gorm:"default:true"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`

	// ✅ NEW FIELDS - Add these for soft delete tracking
	DeletedBy  *string        `gorm:"index"`  // ✅ Use *string to allow NULL
	RestoredAt *time.Time `gorm:"index"`
	RestoredBy *string        `gorm:"index"`  // ✅ Use *string to allow NULL
	IsFeatured bool       `gorm:"default:false;index"`
	IsPrivate  bool       `gorm:"default:false;index"`
}

func (EventModel) TableName() string {
	return "events"
}

// ============================================================
// EVENT TYPE MODEL
// ============================================================

type EventTypeModel struct {
	ID          string `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int `gorm:"default:0"`
	SupportsCertificate bool `gorm:"default:true"`
	MinDuration int `gorm:"default:60"`
	MaxDuration int `gorm:"default:480"`
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (EventTypeModel) TableName() string {
	return "event_types"
}

// ============================================================
// EVENT STATUS MODEL
// ============================================================

type EventStatusModel struct {
	ID          string `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	Color       string
	Icon        string
	SortOrder   int `gorm:"default:0"`
	IsFinal     bool `gorm:"default:false"`
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (EventStatusModel) TableName() string {
	return "event_statuses"
}