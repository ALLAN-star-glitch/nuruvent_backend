// internal/modules/profile/infrastructure/postgres/models.go

package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
	"gorm.io/gorm"
)

// ============================================================
// JSONB TYPE FOR GORM
// ============================================================

// JSONB is a map that can be stored as JSONB in PostgreSQL
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// JSONBString is a map[string]string that can be stored as JSONB
type JSONBString map[string]string

func (j JSONBString) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONBString) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// ============================================================
// USER MODEL
// ============================================================

type UserModel struct {
	ID          string         `gorm:"primaryKey;default:gen_random_uuid()"`
	Name        string         `gorm:"not null"`
	DisplayName string         `gorm:"index"`
	Slug        string         `gorm:"uniqueIndex;not null"`
	Email       string         `gorm:"uniqueIndex;not null"`
	Phone       string         `gorm:"index"`
	AccountType string         `gorm:"index"`
	AvatarURL   string         `gorm:"index"`
	Bio         string
	Location    string         `gorm:"index"`
	Website     string         `gorm:"index"`
	SocialLinks JSONBString    `gorm:"type:jsonb;default:'{}'"`
	IsActive    bool           `gorm:"default:true;index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string {
	return "users"
}

// ============================================================
// ACCOUNT MODEL
// ============================================================

type AccountModel struct {
	ID                   string         `gorm:"primaryKey;default:gen_random_uuid()"`
	Name                 string         `gorm:"not null"`
	DisplayName          string         `gorm:"index"`
	Slug                 string         `gorm:"uniqueIndex;not null"`
	Type                 string         `gorm:"not null;index"` // "personal" or "institution"
	Email                string         `gorm:"uniqueIndex;not null"`
	Phone                string         `gorm:"index"`
	Website              string         `gorm:"index"`
	Description          string
	LogoURL              string
	Address              string
	City                 string         `gorm:"index"`
	Country              string         `gorm:"index"`
	Status               string         `gorm:"default:'active';index"`      // "active", "suspended", "inactive"
	KYCStatus            string         `gorm:"default:'pending';index"`     // "pending", "submitted", "verified", "rejected", "not_required"
	BillingEmail         string
	SubscriptionPlan     string
	BusinessRegistration string
	TaxID                string
	DirectorsDocument    string
	ProofOfAddress       string
	CreatedBy            string         `gorm:"index"`
	IsActive             bool           `gorm:"default:true;index"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            gorm.DeletedAt `gorm:"index"`
}

func (AccountModel) TableName() string {
	return "accounts"
}

// ============================================================
// USER CONVERTERS
// ============================================================

// ToDomain converts UserModel to domain.User
func (m *UserModel) ToDomain() *domain.User {
	if m == nil {
		return nil
	}
	return &domain.User{
		ID:          m.ID,
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Slug:        m.Slug,
		Email:       m.Email,
		Phone:       m.Phone,
		AccountType: m.AccountType,
		AvatarURL:   m.AvatarURL,
		Bio:         m.Bio,
		Location:    m.Location,
		Website:     m.Website,
		SocialLinks: map[string]string(m.SocialLinks),
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// FromDomain converts domain.User to UserModel
func (m *UserModel) FromDomain(user *domain.User) {
	if user == nil {
		return
	}
	m.ID = user.ID
	m.Name = user.Name
	m.DisplayName = user.DisplayName
	m.Slug = user.Slug
	m.Email = user.Email
	m.Phone = user.Phone
	m.AccountType = user.AccountType
	m.AvatarURL = user.AvatarURL
	m.Bio = user.Bio
	m.Location = user.Location
	m.Website = user.Website
	m.SocialLinks = JSONBString(user.SocialLinks)
	m.IsActive = user.IsActive
	m.CreatedAt = user.CreatedAt
	m.UpdatedAt = user.UpdatedAt
}

// ============================================================
// ACCOUNT CONVERTERS
// ============================================================

// ToDomain converts AccountModel to domain.Account
func (m *AccountModel) ToDomain() *domain.Account {
	if m == nil {
		return nil
	}
	return &domain.Account{
		ID:          m.ID,
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Slug:        m.Slug,
		Type:        m.Type,
		Email:       m.Email,
		Phone:       m.Phone,
		Website:     m.Website,
		Description: m.Description,
		LogoURL:     m.LogoURL,
		Address:     m.Address,
		City:        m.City,
		Country:     m.Country,
		Status:      m.Status,
		KYCStatus:   m.KYCStatus,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// FromDomain converts domain.Account to AccountModel
func (m *AccountModel) FromDomain(account *domain.Account) {
	if account == nil {
		return
	}
	m.ID = account.ID
	m.Name = account.Name
	m.DisplayName = account.DisplayName
	m.Slug = account.Slug
	m.Type = account.Type
	m.Email = account.Email
	m.Phone = account.Phone
	m.Website = account.Website
	m.Description = account.Description
	m.LogoURL = account.LogoURL
	m.Address = account.Address
	m.City = account.City
	m.Country = account.Country
	m.Status = account.Status
	m.KYCStatus = account.KYCStatus
	m.IsActive = account.IsActive
	m.CreatedAt = account.CreatedAt
	m.UpdatedAt = account.UpdatedAt
}