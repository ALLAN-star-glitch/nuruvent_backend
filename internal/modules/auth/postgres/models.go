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

	// Relationships
	Users []UserModel `gorm:"foreignKey:AccountTypeID" json:"users,omitempty"`
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
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Users []UserModel `gorm:"foreignKey:ProfessionalTypeID" json:"users,omitempty"`
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
	Icon        string         `gorm:"type:varchar(50)" json:"icon"`
	Color       string         `gorm:"type:varchar(20)" json:"color"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Institutions []InstitutionModel `gorm:"foreignKey:InstitutionTypeID" json:"institutions,omitempty"`
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

	// Relationships
	InstitutionType InstitutionTypeModel `gorm:"foreignKey:InstitutionTypeID" json:"institution_type,omitempty"`
	Users           []UserModel          `gorm:"foreignKey:InstitutionID" json:"users,omitempty"`
	TeamMembers     []TeamMemberModel    `gorm:"foreignKey:InstitutionID" json:"team_members,omitempty"`
}

func (InstitutionModel) TableName() string {
	return "institutions"
}



// ============================================================
// TEAM TYPE MODEL (NEW)
// ============================================================

type TeamTypeModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`          // institution_team, personal_team
	DisplayName string         `gorm:"type:varchar(150);not null" json:"display_name"`   // Institution Team, Personal Team
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`     // institution-team, personal-team
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	TeamMembers []TeamMemberModel `gorm:"foreignKey:TeamTypeID" json:"team_members,omitempty"`
}

func (TeamTypeModel) TableName() string {
	return "team_types"
}

// ============================================================
// TEAM MEMBER MODEL (UPDATED)
// ============================================================

type TeamMemberModel struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MemberID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"member_id"`
	InstitutionID *uuid.UUID     `gorm:"type:uuid;index" json:"institution_id,omitempty"` // NULL for personal teams
	TeamTypeID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"team_type_id"`
	InvitedBy     *uuid.UUID     `gorm:"type:uuid;index" json:"invited_by,omitempty"` // NULL if self-created
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	JoinedAt      time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	CreatedBy     *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Member      UserModel        `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	Institution *InstitutionModel `gorm:"foreignKey:InstitutionID" json:"institution,omitempty"`
	TeamType    TeamTypeModel    `gorm:"foreignKey:TeamTypeID" json:"team_type,omitempty"`
	Inviter     *UserModel       `gorm:"foreignKey:InvitedBy" json:"inviter,omitempty"`
	Creator     *UserModel       `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (TeamMemberModel) TableName() string {
	return "team_members"
}

// ============================================================
// REFRESH TOKEN MODEL
// ============================================================

type RefreshTokenModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Token      string    `gorm:"uniqueIndex;not null;size:512" json:"token"`
	ExpiresAt  time.Time `gorm:"not null" json:"expires_at"`
	Revoked    bool      `gorm:"default:false" json:"revoked"`
	UserAgent  string    `gorm:"size:255" json:"user_agent"`
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`

	// Relationship
	User UserModel `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}


// ============================================================
// USER MODEL (formerly Account)
// ============================================================

type UserModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`
	Email       string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string        `gorm:"not null" json:"-"`
	Phone       string         `gorm:"size:50" json:"phone"`

	AccountTypeID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"account_type_id"`
	ProfessionalTypeID *uuid.UUID `gorm:"type:uuid;index" json:"professional_type_id,omitempty"`
	InstitutionID     *uuid.UUID `gorm:"type:uuid;index" json:"institution_id,omitempty"`

	EmailVerified    bool       `gorm:"default:false" json:"email_verified"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty"`
	IdentityVerified bool       `gorm:"default:false" json:"identity_verified"`
	IdentityVerifiedAt *time.Time `json:"identity_verified_at,omitempty"`
	PhoneVerified    bool       `gorm:"default:false" json:"phone_verified"`
	PhoneVerifiedAt  *time.Time `json:"phone_verified_at,omitempty"`
	KYCStatus        string     `gorm:"type:varchar(50);default:'pending'" json:"kyc_status"`

	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	AccountType      AccountTypeModel        `gorm:"foreignKey:AccountTypeID" json:"account_type,omitempty"`
	ProfessionalType *ProfessionalTypeModel  `gorm:"foreignKey:ProfessionalTypeID" json:"professional_type,omitempty"`
	Institution      *InstitutionModel       `gorm:"foreignKey:InstitutionID" json:"institution,omitempty"`
	TeamMembers      []TeamMemberModel       `gorm:"foreignKey:MemberID" json:"team_members,omitempty"`
	RefreshTokens    []RefreshTokenModel     `gorm:"foreignKey:UserID" json:"refresh_tokens,omitempty"`
}

func (UserModel) TableName() string {
	return "users"
}

