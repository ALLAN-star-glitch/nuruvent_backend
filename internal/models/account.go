// internal/models/account.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Account represents a user account in the system
type Account struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`
	Email       string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Password    string         `gorm:"not null" json:"-"`
	Phone       string         `gorm:"size:50" json:"phone"`
	
	// Foreign Keys
	AccountTypeID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"account_type_id"`
	ProfessionalTypeID *uuid.UUID `gorm:"type:uuid;index" json:"professional_type_id,omitempty"`
	InstitutionID     *uuid.UUID `gorm:"type:uuid;index" json:"institution_id,omitempty"`
	
	// Verification Status
	EmailVerified    bool       `gorm:"default:false" json:"email_verified"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty"`
	IdentityVerified bool       `gorm:"default:false" json:"identity_verified"`
	IdentityVerifiedAt *time.Time `json:"identity_verified_at,omitempty"`
	
	// Status
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Essential Relationships Only
	AccountType      AccountType       `gorm:"foreignKey:AccountTypeID" json:"account_type,omitempty"`
	ProfessionalType *ProfessionalType `gorm:"foreignKey:ProfessionalTypeID" json:"professional_type,omitempty"`
	Institution      *Institution      `gorm:"foreignKey:InstitutionID" json:"institution,omitempty"`
	TeamMembers      []TeamMember      `gorm:"foreignKey:AccountID" json:"team_members,omitempty"`
}