package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BusinessType struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(50);unique;not null" json:"name"`  // training_institute, college, individual_formal, individual_informal, etc.
	DisplayName string         `gorm:"type:varchar(100);not null" json:"display_name"` // Training Institute
	Description string         `gorm:"type:text" json:"description"`
	Icon        string         `gorm:"type:varchar(50)" json:"icon"` // school, university, briefcase, user, etc.
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	
	// Business type category for grouping
	Category    string         `gorm:"type:varchar(50);default:'organization'" json:"category"` // individual, organization
	
	// For individual types: indicates if business registration is required
	RequiresBusinessName bool   `gorm:"default:false" json:"requires_business_name"`
	RequiresBusinessEmail bool  `gorm:"default:false" json:"requires_business_email"`
	
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BusinessType) TableName() string {
	return "business_types"
}