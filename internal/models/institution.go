// internal/models/institution.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Institution represents an organization/company/institute
type Institution struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`
	Name        string         `gorm:"not null;size:255" json:"name"`
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`
	Email       string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Phone       string         `gorm:"size:50" json:"phone"`
	
	// Institution Type (Foreign Key)
	InstitutionTypeID uuid.UUID `gorm:"type:uuid;index;not null" json:"institution_type_id"`
	
	Description string         `gorm:"type:text" json:"description"`
	Logo        string         `gorm:"size:500" json:"logo"`
	Website     string         `gorm:"size:255" json:"website"`
	Address     string         `gorm:"type:text" json:"address"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	InstitutionType InstitutionType `gorm:"foreignKey:InstitutionTypeID" json:"institution_type,omitempty"`
	Accounts        []Account       `gorm:"foreignKey:InstitutionID" json:"accounts,omitempty"`
}