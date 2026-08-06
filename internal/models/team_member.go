// internal/models/team_member.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamMemberRole string

const (
	TeamMemberRoleAdmin       TeamMemberRole = "admin"
	TeamMemberRoleEventManager TeamMemberRole = "event_manager"
	TeamMemberRoleTeamMember  TeamMemberRole = "team_member"
)

type TeamMember struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug          string         `gorm:"type:varchar(50);unique;not null" json:"slug"`
	Name          string         `gorm:"type:varchar(100);not null" json:"name"`
	DisplayName   string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`
	AccountID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"account_id"`
	MemberID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"member_id"`
	Role          TeamMemberRole `gorm:"type:varchar(50);not null;default:'team_member'" json:"role"`
	JobTitle      string         `gorm:"size:100" json:"job_title,omitempty"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	CreatedBy     uuid.UUID      `gorm:"type:uuid" json:"created_by,omitempty"`
	JoinedAt      time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Account Account `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	Member  Account `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	Creator Account `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}