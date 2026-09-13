// internal/modules/auth/infrastructure/postgres/models.go

package postgres

import (
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"gorm.io/gorm"
)

// ============================================================
// ACCOUNT TYPE MODEL
// ============================================================

type AccountTypeModel struct {
	ID          string         `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug        string         `gorm:"type:varchar(50);uniqueIndex;not null"`
	Name        string         `gorm:"type:varchar(100);not null"`
	DisplayName string         `gorm:"type:varchar(150)"`
	Description string         `gorm:"type:text"`
	Icon        string         `gorm:"type:varchar(50)"`
	Color       string         `gorm:"type:varchar(20)"`
	SortOrder   int            `gorm:"default:0"`
	IsActive    bool           `gorm:"default:true"`
	CreatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (AccountTypeModel) TableName() string {
	return "account_types"
}

func (m *AccountTypeModel) ToDomain() *authdomain.AccountType {
	if m == nil {
		return nil
	}
	return &authdomain.AccountType{
		ID:          m.ID,
		Slug:        m.Slug,
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Description: m.Description,
		Icon:        m.Icon,
		Color:       m.Color,
		SortOrder:   m.SortOrder,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   &m.DeletedAt.Time,
	}
}

func (m *AccountTypeModel) FromDomain(at *authdomain.AccountType) {
	if at == nil {
		return
	}
	m.ID = at.ID
	m.Slug = at.Slug
	m.Name = at.Name
	m.DisplayName = at.DisplayName
	m.Description = at.Description
	m.Icon = at.Icon
	m.Color = at.Color
	m.SortOrder = at.SortOrder
	m.IsActive = at.IsActive
	m.CreatedAt = at.CreatedAt
	m.UpdatedAt = at.UpdatedAt
	if at.DeletedAt != nil {
		m.DeletedAt = gorm.DeletedAt{Time: *at.DeletedAt, Valid: true}
	}
}

// ============================================================
// PROFESSIONAL TYPE MODEL
// ============================================================

type ProfessionalTypeModel struct {
	ID          string         `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug        string         `gorm:"type:varchar(50);uniqueIndex;not null"`
	Name        string         `gorm:"type:varchar(100);not null"`
	DisplayName string         `gorm:"type:varchar(150)"`
	Description string         `gorm:"type:text"`
	Icon        string         `gorm:"type:varchar(50)"`
	Color       string         `gorm:"type:varchar(20)"`
	SortOrder   int            `gorm:"default:0"`
	IsActive    bool           `gorm:"default:true"`
	CreatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (ProfessionalTypeModel) TableName() string {
	return "professional_types"
}

func (m *ProfessionalTypeModel) ToDomain() *authdomain.ProfessionalType {
	if m == nil {
		return nil
	}
	return &authdomain.ProfessionalType{
		ID:          m.ID,
		Slug:        m.Slug,
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Description: m.Description,
		Icon:        m.Icon,
		Color:       m.Color,
		SortOrder:   m.SortOrder,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   &m.DeletedAt.Time,
	}
}

func (m *ProfessionalTypeModel) FromDomain(pt *authdomain.ProfessionalType) {
	if pt == nil {
		return
	}
	m.ID = pt.ID
	m.Slug = pt.Slug
	m.Name = pt.Name
	m.DisplayName = pt.DisplayName
	m.Description = pt.Description
	m.Icon = pt.Icon
	m.Color = pt.Color
	m.SortOrder = pt.SortOrder
	m.IsActive = pt.IsActive
	m.CreatedAt = pt.CreatedAt
	m.UpdatedAt = pt.UpdatedAt
	if pt.DeletedAt != nil {
		m.DeletedAt = gorm.DeletedAt{Time: *pt.DeletedAt, Valid: true}
	}
}

// ============================================================
// INSTITUTION TYPE MODEL
// ============================================================

type InstitutionTypeModel struct {
	ID          string         `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug        string         `gorm:"type:varchar(50);uniqueIndex;not null"`
	Name        string         `gorm:"type:varchar(100);not null"`
	DisplayName string         `gorm:"type:varchar(150)"`
	Description string         `gorm:"type:text"`
	Icon        string         `gorm:"type:varchar(50)"`
	Color       string         `gorm:"type:varchar(20)"`
	SortOrder   int            `gorm:"default:0"`
	IsActive    bool           `gorm:"default:true"`
	CreatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (InstitutionTypeModel) TableName() string {
	return "institution_types"
}

func (m *InstitutionTypeModel) ToDomain() *authdomain.InstitutionType {
	if m == nil {
		return nil
	}
	return &authdomain.InstitutionType{
		ID:          m.ID,
		Slug:        m.Slug,
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Description: m.Description,
		Icon:        m.Icon,
		Color:       m.Color,
		SortOrder:   m.SortOrder,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   &m.DeletedAt.Time,
	}
}

func (m *InstitutionTypeModel) FromDomain(it *authdomain.InstitutionType) {
	if it == nil {
		return
	}
	m.ID = it.ID
	m.Slug = it.Slug
	m.Name = it.Name
	m.DisplayName = it.DisplayName
	m.Description = it.Description
	m.Icon = it.Icon
	m.Color = it.Color
	m.SortOrder = it.SortOrder
	m.IsActive = it.IsActive
	m.CreatedAt = it.CreatedAt
	m.UpdatedAt = it.UpdatedAt
	if it.DeletedAt != nil {
		m.DeletedAt = gorm.DeletedAt{Time: *it.DeletedAt, Valid: true}
	}
}

// ============================================================
// USER MODEL
// ============================================================

type UserModel struct {
	ID                  string         `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug                string         `gorm:"type:varchar(50);uniqueIndex;not null"`
	Name                string         `gorm:"type:varchar(100);not null"`
	DisplayName         string         `gorm:"type:varchar(150)"`
	Email               string         `gorm:"uniqueIndex;not null;size:255"`
	PasswordHash        string         `gorm:"not null"`
	Phone               string         `gorm:"size:50"`
	AccountTypeID       string         `gorm:"type:uuid;index;not null"`
	ProfessionalTypeID  *string        `gorm:"type:uuid;index"`
	EmailVerified       bool           `gorm:"default:false"`
	EmailVerifiedAt     *time.Time
	IdentityVerified    bool           `gorm:"default:false"`
	IdentityVerifiedAt  *time.Time
	PhoneVerified       bool           `gorm:"default:false"`
	PhoneVerifiedAt     *time.Time
	IsActive            bool           `gorm:"default:true"`
	CreatedAt           time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt           time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt           gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string {
	return "users"
}

func (m *UserModel) ToDomain() *authdomain.User {
	if m == nil {
		return nil
	}
	return &authdomain.User{
		ID:                  m.ID,
		Slug:                m.Slug,
		Name:                m.Name,
		DisplayName:         m.DisplayName,
		Email:               m.Email,
		PasswordHash:        m.PasswordHash,
		Phone:               m.Phone,
		AccountTypeID:       m.AccountTypeID,
		ProfessionalTypeID:  m.ProfessionalTypeID,
		EmailVerified:       m.EmailVerified,
		EmailVerifiedAt:     m.EmailVerifiedAt,
		IdentityVerified:    m.IdentityVerified,
		IdentityVerifiedAt:  m.IdentityVerifiedAt,
		PhoneVerified:       m.PhoneVerified,
		PhoneVerifiedAt:     m.PhoneVerifiedAt,
		IsActive:            m.IsActive,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
		DeletedAt:           &m.DeletedAt.Time,
	}
}

func (m *UserModel) FromDomain(user *authdomain.User) {
	if user == nil {
		return
	}
	m.ID = user.ID
	m.Slug = user.Slug
	m.Name = user.Name
	m.DisplayName = user.DisplayName
	m.Email = user.Email
	m.PasswordHash = user.PasswordHash
	m.Phone = user.Phone
	m.AccountTypeID = user.AccountTypeID
	m.ProfessionalTypeID = user.ProfessionalTypeID
	m.EmailVerified = user.EmailVerified
	m.EmailVerifiedAt = user.EmailVerifiedAt
	m.IdentityVerified = user.IdentityVerified
	m.IdentityVerifiedAt = user.IdentityVerifiedAt
	m.PhoneVerified = user.PhoneVerified
	m.PhoneVerifiedAt = user.PhoneVerifiedAt
	m.IsActive = user.IsActive
	m.CreatedAt = user.CreatedAt
	m.UpdatedAt = user.UpdatedAt
	if user.DeletedAt != nil {
		m.DeletedAt = gorm.DeletedAt{Time: *user.DeletedAt, Valid: true}
	}
}

// ============================================================
// ACCOUNT MODEL
// ============================================================

type AccountModel struct {
	ID                string         `gorm:"primaryKey;default:gen_random_uuid()"`
	Name              string         `gorm:"column:name;type:varchar(255);not null"`
	DisplayName       string         `gorm:"column:display_name;type:varchar(255);not null"`
	Slug              string         `gorm:"column:slug;type:varchar(255);uniqueIndex;not null"`
	Email             string         `gorm:"column:email;type:varchar(255);uniqueIndex;not null"`
	Phone             string         `gorm:"column:phone;type:varchar(50)"`
	AccountTypeID     string         `gorm:"column:account_type_id;type:uuid;not null;index"`
	InstitutionTypeID *string        `gorm:"column:institution_type_id;type:uuid;index"`
	Status            string         `gorm:"column:status;type:varchar(50);default:'active'"`
	LogoURL           string         `gorm:"column:logo_url;type:varchar(500)"`
	Website           string         `gorm:"column:website;type:varchar(255)"`
	Description       string         `gorm:"column:description;type:text"`
	Address           string         `gorm:"column:address;type:text"`
	City              string         `gorm:"column:city;type:varchar(100)"`
	Country           string         `gorm:"column:country;type:varchar(100)"`
	BillingEmail      string         `gorm:"column:billing_email;type:varchar(255)"`
	SubscriptionPlan  string         `gorm:"column:subscription_plan;type:varchar(50)"`
	CreatedBy         string         `gorm:"column:created_by;type:uuid;index"`
	CreatedAt         time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	DeletedAt         *time.Time     `gorm:"column:deleted_at;index"`
}

func (AccountModel) TableName() string {
	return "accounts"
}

func (m *AccountModel) ToDomain() *authdomain.Account {
	if m == nil {
		return nil
	}
	return &authdomain.Account{
		ID:                m.ID,
		Name:              m.Name,
		DisplayName:       m.DisplayName,
		Slug:              m.Slug,
		Email:             m.Email,
		Phone:             m.Phone,
		AccountTypeID:     m.AccountTypeID,
		InstitutionTypeID: m.InstitutionTypeID,
		Status:            m.Status,
		LogoURL:           m.LogoURL,
		Website:           m.Website,
		Description:       m.Description,
		Address:           m.Address,
		City:              m.City,
		Country:           m.Country,
		BillingEmail:      m.BillingEmail,
		SubscriptionPlan:  m.SubscriptionPlan,
		CreatedBy:         m.CreatedBy,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
		DeletedAt:         m.DeletedAt,
	}
}

func (m *AccountModel) FromDomain(account *authdomain.Account) {
	if account == nil {
		return
	}
	m.ID = account.ID
	m.Name = account.Name
	m.DisplayName = account.DisplayName
	m.Slug = account.Slug
	m.Email = account.Email
	m.Phone = account.Phone
	m.AccountTypeID = account.AccountTypeID
	m.InstitutionTypeID = account.InstitutionTypeID
	m.Status = account.Status
	m.LogoURL = account.LogoURL
	m.Website = account.Website
	m.Description = account.Description
	m.Address = account.Address
	m.City = account.City
	m.Country = account.Country
	m.BillingEmail = account.BillingEmail
	m.SubscriptionPlan = account.SubscriptionPlan
	m.CreatedBy = account.CreatedBy
	m.CreatedAt = account.CreatedAt
	m.UpdatedAt = account.UpdatedAt
	m.DeletedAt = account.DeletedAt
}

// ============================================================
// ACCOUNT MEMBER MODEL
// ============================================================

type AccountMemberModel struct {
	ID          string         `gorm:"primaryKey;default:gen_random_uuid()"`
	AccountID   string         `gorm:"column:account_id;type:uuid;not null;index"`
	UserID      string         `gorm:"column:user_id;type:uuid;not null;index"`
	Role        string         `gorm:"column:role;type:varchar(50);not null"`
	IsActive    bool           `gorm:"column:is_active;default:true"`
	InvitedBy   string         `gorm:"column:invited_by;type:uuid;index"`
	JoinedAt    time.Time      `gorm:"column:joined_at;default:CURRENT_TIMESTAMP"`
	CreatedAt   time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	DeletedAt   *time.Time     `gorm:"column:deleted_at;index"`
}

func (AccountMemberModel) TableName() string {
	return "account_members"
}

func (m *AccountMemberModel) ToDomain() *authdomain.AccountMember {
	if m == nil {
		return nil
	}
	return &authdomain.AccountMember{
		ID:          m.ID,
		AccountID:   m.AccountID,
		UserID:      m.UserID,
		Role:        m.Role,
		IsActive:    m.IsActive,
		InvitedBy:   m.InvitedBy,
		JoinedAt:    m.JoinedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
	}
}

func (m *AccountMemberModel) FromDomain(member *authdomain.AccountMember) {
	if member == nil {
		return
	}
	m.ID = member.ID
	m.AccountID = member.AccountID
	m.UserID = member.UserID
	m.Role = member.Role
	m.IsActive = member.IsActive
	m.InvitedBy = member.InvitedBy
	m.JoinedAt = member.JoinedAt
	m.CreatedAt = member.CreatedAt
	m.UpdatedAt = member.UpdatedAt
	m.DeletedAt = member.DeletedAt
}

// ============================================================
// REFRESH TOKEN MODEL
// ============================================================

type RefreshTokenModel struct {
	ID         string     `gorm:"primaryKey;default:gen_random_uuid()"`
	UserID     string     `gorm:"type:uuid;index;not null"`
	Token      string     `gorm:"uniqueIndex;not null;size:512"`
	ExpiresAt  time.Time  `gorm:"not null"`
	Revoked    bool       `gorm:"default:false"`
	UserAgent  string     `gorm:"size:255"`
	IPAddress  string     `gorm:"size:45"`
	CreatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt  *time.Time `gorm:"index"`
}

func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

func (m *RefreshTokenModel) ToDomain() *authdomain.RefreshToken {
	if m == nil {
		return nil
	}
	return &authdomain.RefreshToken{
		ID:         m.ID,
		UserID:     m.UserID,
		Token:      m.Token,
		ExpiresAt:  m.ExpiresAt,
		Revoked:    m.Revoked,
		UserAgent:  m.UserAgent,
		IPAddress:  m.IPAddress,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
		DeletedAt:  m.DeletedAt,
	}
}

func (m *RefreshTokenModel) FromDomain(rt *authdomain.RefreshToken) {
	if rt == nil {
		return
	}
	m.ID = rt.ID
	m.UserID = rt.UserID
	m.Token = rt.Token
	m.ExpiresAt = rt.ExpiresAt
	m.Revoked = rt.Revoked
	m.UserAgent = rt.UserAgent
	m.IPAddress = rt.IPAddress
	m.CreatedAt = rt.CreatedAt
	m.UpdatedAt = rt.UpdatedAt
	m.DeletedAt = rt.DeletedAt
}