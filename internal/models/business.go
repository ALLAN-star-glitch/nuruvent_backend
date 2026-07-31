package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Business struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string         `gorm:"not null" json:"name"`
	Slug           string         `gorm:"uniqueIndex;not null" json:"slug"`
	BusinessTypeID uuid.UUID      `gorm:"type:uuid;index;not null" json:"business_type_id"`
	Email          string         `gorm:"uniqueIndex" json:"email,omitempty"`
	Phone          string         `json:"phone,omitempty"`
	Address        string         `json:"address,omitempty"`
	Logo           string         `gorm:"type:varchar(500)" json:"logo,omitempty"`
	Website        string         `gorm:"type:varchar(255)" json:"website,omitempty"`
	Description    string         `gorm:"type:text" json:"description,omitempty"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	IsEmailVerified bool          `gorm:"default:false" json:"is_email_verified"`
	EmailVerifiedAt *time.Time    `json:"email_verified_at,omitempty"`
	IsVerified     bool           `gorm:"default:false" json:"is_verified"`
	VerifiedAt     *time.Time     `json:"verified_at,omitempty"`
	VerifiedBy     *uuid.UUID     `gorm:"type:uuid" json:"verified_by,omitempty"`
	VerificationMethod string     `gorm:"type:varchar(50)" json:"verification_method,omitempty"`
	CreatedBy      uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	BusinessType *BusinessType    `gorm:"foreignKey:BusinessTypeID" json:"business_type,omitempty"`
	Members      []BusinessMember `gorm:"foreignKey:BusinessID" json:"members,omitempty"`
	Events       []Event          `gorm:"foreignKey:BusinessID" json:"events,omitempty"`
}

func (Business) TableName() string {
	return "businesses"
}