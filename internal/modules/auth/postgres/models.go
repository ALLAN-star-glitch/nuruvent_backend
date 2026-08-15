package postgres

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================
// ACCOUNT TYPE MODEL
// ============================================================

type AccountTypeModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`
	Description string         `gorm:"type:text" json:"description"`
	Icon        string         `gorm:"type:varchar(50)" json:"icon"`
	Color       string         `gorm:"type:varchar(20)" json:"color"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// ✅ Relationship WITHIN auth module
	Accounts []AccountModel `gorm:"foreignKey:AccountTypeID" json:"accounts,omitempty"`
}

func (AccountTypeModel) TableName() string {
	return "account_types"
}

// ============================================================
// PROFESSIONAL TYPE MODEL
// ============================================================

type ProfessionalTypeModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`
	Description string         `gorm:"type:text" json:"description"`
	Icon        string         `gorm:"type:varchar(50)" json:"icon"`
	Color       string         `gorm:"type:varchar(20)" json:"color"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CanHost     bool           `gorm:"default:false" json:"can_host"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// ✅ Relationship WITHIN auth module
	Accounts []AccountModel `gorm:"foreignKey:ProfessionalTypeID" json:"accounts,omitempty"`
}

func (ProfessionalTypeModel) TableName() string {
	return "professional_types"
}

// ============================================================
// INSTITUTION TYPE MODEL
// ============================================================

type InstitutionTypeModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (InstitutionTypeModel) TableName() string {
	return "institution_types"
}

// ============================================================
// INSTITUTION MODEL
// ============================================================

type InstitutionModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`
	Name        string         `gorm:"not null;size:255" json:"name"`
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`
	Email       string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Phone       string         `gorm:"size:50" json:"phone"`
	
	InstitutionTypeID uuid.UUID `gorm:"type:uuid;index;not null" json:"institution_type_id"`
	
	Description string         `gorm:"type:text" json:"description"`
	Logo        string         `gorm:"size:500" json:"logo"`
	Website     string         `gorm:"size:255" json:"website"`
	Address     string         `gorm:"type:text" json:"address"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// ✅ Relationships WITHIN auth module
	InstitutionType InstitutionTypeModel `gorm:"foreignKey:InstitutionTypeID" json:"institution_type,omitempty"`
	Accounts        []AccountModel       `gorm:"foreignKey:InstitutionID" json:"accounts,omitempty"`
}

func (InstitutionModel) TableName() string {
	return "institutions"
}

// ============================================================
// TEAM MEMBER MODEL
// ============================================================

type TeamMemberRole string

const (
	TeamMemberRoleAdmin        TeamMemberRole = "account_admin"
	TeamMemberRoleEventManager TeamMemberRole = "event_manager"
	TeamMemberRoleTeamMember   TeamMemberRole = "team_member"
)

type TeamMemberModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`
	AccountID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"account_id"`
	MemberID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"member_id"`
	Role        TeamMemberRole `gorm:"type:varchar(50);not null;default:'team_member'" json:"role"`
	JobTitle    string         `gorm:"size:100" json:"job_title,omitempty"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
    CreatedBy   *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	JoinedAt    time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// ✅ Relationships WITHIN auth module
	Account AccountModel `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	Member  AccountModel `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	Creator AccountModel `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (TeamMemberModel) TableName() string {
	return "team_members"
}

// ============================================================
// REFRESH TOKEN MODEL
// ============================================================

type RefreshTokenModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AccountID  uuid.UUID `gorm:"type:uuid;index;not null" json:"account_id"`
	Token      string    `gorm:"uniqueIndex;not null;size:512" json:"token"`
	ExpiresAt  time.Time `gorm:"not null" json:"expires_at"`
	Revoked    bool      `gorm:"default:false" json:"revoked"`
	UserAgent  string    `gorm:"size:255" json:"user_agent"`
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// ✅ Relationship WITHIN auth module
	Account AccountModel `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}

func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

// ============================================================
// ACCOUNT MODEL (Main Entity)
// ============================================================

type AccountModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`
	Email       string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash    string         `gorm:"not null" json:"-"`
	Phone       string         `gorm:"size:50" json:"phone"`
	
	AccountTypeID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"account_type_id"`
	ProfessionalTypeID *uuid.UUID `gorm:"type:uuid;index" json:"professional_type_id,omitempty"`
	InstitutionID     *uuid.UUID `gorm:"type:uuid;index" json:"institution_id,omitempty"`
	
	EmailVerified    bool       `gorm:"default:false" json:"email_verified"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty"`
	IdentityVerified bool       `gorm:"default:false" json:"identity_verified"`
	IdentityVerifiedAt *time.Time `json:"identity_verified_at,omitempty"`
	
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// ✅ Relationships WITHIN auth module ONLY
	AccountType      AccountTypeModel       `gorm:"foreignKey:AccountTypeID" json:"account_type,omitempty"`
	ProfessionalType *ProfessionalTypeModel `gorm:"foreignKey:ProfessionalTypeID" json:"professional_type,omitempty"`
	Institution      *InstitutionModel      `gorm:"foreignKey:InstitutionID" json:"institution,omitempty"`
	TeamMembers      []TeamMemberModel      `gorm:"foreignKey:AccountID" json:"team_members,omitempty"`
}

func (AccountModel) TableName() string {
	return "accounts"
}