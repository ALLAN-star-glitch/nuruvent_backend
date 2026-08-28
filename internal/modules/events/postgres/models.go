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
	InstitutionID    string `gorm:"index"` // ✅ Changed from AccountID
	CreatedBy        string `gorm:"index"` // User ID who created this event
	IsActive         bool   `gorm:"default:true"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`

	// Soft delete tracking
	DeletedBy  *string        `gorm:"index"`
	RestoredAt *time.Time     `gorm:"index"`
	RestoredBy *string        `gorm:"index"`
	IsFeatured bool           `gorm:"default:false;index"`
	IsPrivate  bool           `gorm:"default:false;index"`

	// Creator info (populated via JOIN for display purposes only)
	CreatorName            string `gorm:"column:creator_name;<-:false"`
	CreatorDisplayName     string `gorm:"column:creator_display_name;<-:false"`
	CreatorEmail           string `gorm:"column:creator_email;<-:false"`
	CreatorPhone           string `gorm:"column:creator_phone;<-:false"`
	CreatorAccountType     string `gorm:"column:creator_account_type;<-:false"`
	CreatorInstitutionName string `gorm:"column:creator_institution_name;<-:false"`
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

// ============================================================
// USER MODEL (for creator info)
// ============================================================

type UserModel struct {
	ID               string `gorm:"primaryKey"`
	Name             string
	DisplayName      string
	Email            string
	Phone            string
	AccountTypeID    string `gorm:"index"`
	InstitutionID    *string `gorm:"index"`
	InstitutionName  string `gorm:"-"` // Not a DB column, populated via JOIN
	IsActive         bool   `gorm:"default:true"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string {
	return "users"
}

// ============================================================
// INSTITUTION MODEL
// ============================================================

type InstitutionModel struct {
	ID          string `gorm:"primaryKey"`
	Name        string
	DisplayName string
	Slug        string
	Email       string
	Phone       string
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (InstitutionModel) TableName() string {
	return "institutions"
}

// ============================================================
// ACCOUNT TYPE MODEL
// ============================================================

type AccountTypeModel struct {
	ID          string `gorm:"primaryKey"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (AccountTypeModel) TableName() string {
	return "account_types"
}