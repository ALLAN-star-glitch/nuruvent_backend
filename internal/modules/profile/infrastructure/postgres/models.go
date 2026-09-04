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

// ============================================================
// JSONBString is a map[string]string for JSONB
// ============================================================

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
    DisplayName string
    Email       string         `gorm:"uniqueIndex;not null"`
    Phone       string
    AccountType string         `gorm:"index"`
    AvatarURL   string
    Bio         string
    Location    string
    Website     string
    SocialLinks JSONBString    `gorm:"type:jsonb;default:'{}'"` // ✅ Use JSONBString
    IsActive    bool           `gorm:"default:true;index"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string {
    return "users"
}

// ============================================================
// INSTITUTION MODEL
// ============================================================

type InstitutionModel struct {
    ID          string         `gorm:"primaryKey;default:gen_random_uuid()"`
    Name        string         `gorm:"not null"`
    DisplayName string
    Slug        string         `gorm:"uniqueIndex;not null"`
    Email       string
    Phone       string
    Website     string
    Description string
    LogoURL     string
    Address     string
    City        string
    Country     string
    IsActive    bool           `gorm:"default:true;index"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (InstitutionModel) TableName() string {
    return "institutions"
}

// ============================================================
// CONVERTERS
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
        Email:       m.Email,
        Phone:       m.Phone,
        AvatarURL:   m.AvatarURL,
        Bio:         m.Bio,
        Location:    m.Location,
        Website:     m.Website,
        SocialLinks: map[string]string(m.SocialLinks), // ✅ Convert JSONBString to map[string]string
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
    m.Email = user.Email
    m.Phone = user.Phone
    // ❌ REMOVE: m.AccountType = user.AccountType
    // ✅ This should be handled separately if needed
    m.AvatarURL = user.AvatarURL
    m.Bio = user.Bio
    m.Location = user.Location
    m.Website = user.Website
    m.SocialLinks = JSONBString(user.SocialLinks)
    m.IsActive = user.IsActive
    m.CreatedAt = user.CreatedAt
    m.UpdatedAt = user.UpdatedAt
}

// ToDomain converts InstitutionModel to domain.Institution
func (m *InstitutionModel) ToDomain() *domain.Institution {
    if m == nil {
        return nil
    }
    return &domain.Institution{
        ID:          m.ID,
        Name:        m.Name,
        DisplayName: m.DisplayName,
        Slug:        m.Slug,
        Email:       m.Email,
        Phone:       m.Phone,
        Website:     m.Website,
        Description: m.Description,
        LogoURL:     m.LogoURL,
        Address:     m.Address,
        City:        m.City,
        Country:     m.Country,
        IsActive:    m.IsActive,
        CreatedAt:   m.CreatedAt,
        UpdatedAt:   m.UpdatedAt,
    }
}

// FromDomain converts domain.Institution to InstitutionModel
func (m *InstitutionModel) FromDomain(institution *domain.Institution) {
    if institution == nil {
        return
    }
    m.ID = institution.ID
    m.Name = institution.Name
    m.DisplayName = institution.DisplayName
    m.Slug = institution.Slug
    m.Email = institution.Email
    m.Phone = institution.Phone
    m.Website = institution.Website
    m.Description = institution.Description
    m.LogoURL = institution.LogoURL
    m.Address = institution.Address
    m.City = institution.City
    m.Country = institution.Country
    m.IsActive = institution.IsActive
    m.CreatedAt = institution.CreatedAt
    m.UpdatedAt = institution.UpdatedAt
}