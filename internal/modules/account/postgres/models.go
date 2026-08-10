package postgres

import (
	"time"

	"gorm.io/gorm"
)

// ============================================================
// ACCOUNT MODEL (GORM)
// ============================================================

type AccountModel struct {
	ID             string `gorm:"primaryKey"`
	Slug           string `gorm:"uniqueIndex"`
	Name           string
	DisplayName    string
	Email          string `gorm:"uniqueIndex"`
	Password       string
	Phone          string `gorm:"uniqueIndex"`
	AccountTypeID  string `gorm:"index"`
	ProfessionalTypeID *string
	InstitutionID  *string
	EmailVerified  bool
	EmailVerifiedAt *time.Time
	IdentityVerified bool
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (AccountModel) TableName() string {
	return "accounts"
}

// ============================================================
// ACCOUNT TYPE MODEL (GORM)
// ============================================================

type AccountTypeModel struct {
	ID          string `gorm:"primaryKey"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int    `gorm:"default:0"`
	IsActive    bool   `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (AccountTypeModel) TableName() string {
	return "account_types"
}

// ============================================================
// PROFESSIONAL TYPE MODEL (GORM)
// ============================================================

type ProfessionalTypeModel struct {
	ID          string `gorm:"primaryKey"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int    `gorm:"default:0"`
	CanHost     bool   `gorm:"default:false"`
	IsActive    bool   `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (ProfessionalTypeModel) TableName() string {
	return "professional_types"
}

// ============================================================
// INSTITUTION TYPE MODEL (GORM)
// ============================================================

type InstitutionTypeModel struct {
	ID          string `gorm:"primaryKey"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int    `gorm:"default:0"`
	IsActive    bool   `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (InstitutionTypeModel) TableName() string {
	return "institution_types"
}

// ============================================================
// INSTITUTION MODEL (GORM)
// ============================================================

type InstitutionModel struct {
	ID                string `gorm:"primaryKey"`
	Slug              string `gorm:"uniqueIndex"`
	Name              string `gorm:"not null"`
	DisplayName       string
	Email             string `gorm:"uniqueIndex"`
	Phone             string
	InstitutionTypeID string `gorm:"index"`
	Description       string
	Logo              string
	Website           string
	Address           string
	IsActive          bool `gorm:"default:true"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

func (InstitutionModel) TableName() string {
	return "institutions"
}

// ============================================================
// TEAM MEMBER MODEL (GORM)
// ============================================================

type TeamMemberModel struct {
	ID         string `gorm:"primaryKey"`
	Slug       string `gorm:"uniqueIndex"`
	Name       string `gorm:"not null"`
	DisplayName string
	AccountID  string `gorm:"index;not null"`
	MemberID   string `gorm:"index;not null"`
	Role       string `gorm:"default:'team_member'"`
	JobTitle   string
	IsActive   bool `gorm:"default:true"`
	JoinedAt   time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (TeamMemberModel) TableName() string {
	return "team_members"
}