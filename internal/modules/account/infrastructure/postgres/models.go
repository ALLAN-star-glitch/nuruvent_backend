// internal/modules/account/infrastructure/postgres/models.go

package postgres

import (
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// ACCOUNT TYPE MODEL
// ============================================================

type AccountTypeModel struct {
	ID          string     `gorm:"primaryKey;default:gen_random_uuid()"`
	Name        string     `gorm:"column:name;type:varchar(100);uniqueIndex;not null"`
	DisplayName string     `gorm:"column:display_name;type:varchar(100);not null"`
	Slug        string     `gorm:"column:slug;type:varchar(100);uniqueIndex;not null"`
	Description string     `gorm:"column:description;type:text"`
	Icon        string     `gorm:"column:icon;type:varchar(50)"`
	Color       string     `gorm:"column:color;type:varchar(20)"`
	SortOrder   int        `gorm:"column:sort_order;default:0"`
	IsActive    bool       `gorm:"column:is_active;default:true"`
	CreatedAt   time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;index"`
}

func (AccountTypeModel) TableName() string {
	return "account_types"
}

func (m *AccountTypeModel) ToDomain() *accountdomain.AccountType {
	if m == nil {
		return nil
	}
	return &accountdomain.AccountType{
		ID:          m.ID,
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Slug:        m.Slug,
		Description: m.Description,
		Icon:        m.Icon,
		Color:       m.Color,
		SortOrder:   m.SortOrder,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
	}
}

func (m *AccountTypeModel) FromDomain(at *accountdomain.AccountType) {
	if at == nil {
		return
	}
	m.ID = at.ID
	m.Name = at.Name
	m.DisplayName = at.DisplayName
	m.Slug = at.Slug
	m.Description = at.Description
	m.Icon = at.Icon
	m.Color = at.Color
	m.SortOrder = at.SortOrder
	m.IsActive = at.IsActive
	m.CreatedAt = at.CreatedAt
	m.UpdatedAt = at.UpdatedAt
	m.DeletedAt = at.DeletedAt
}

// ============================================================
// ACCOUNT MODEL
// ============================================================

type AccountModel struct {
	ID            string     `gorm:"primaryKey;default:gen_random_uuid()"`
	Name          string     `gorm:"column:name;type:varchar(255);not null"`
	DisplayName   string     `gorm:"column:display_name;type:varchar(255);not null"`
	Slug          string     `gorm:"column:slug;type:varchar(255);uniqueIndex;not null"`
	Email         string     `gorm:"column:email;type:varchar(255);uniqueIndex;not null"`
	Phone         string     `gorm:"column:phone;type:varchar(50)"`
	AccountTypeID string     `gorm:"column:account_type_id;type:uuid;not null;index"`
	Status        string     `gorm:"column:status;type:varchar(50);default:'active'"`
	LogoURL       string     `gorm:"column:logo_url;type:varchar(500)"`
	Website       string     `gorm:"column:website;type:varchar(255)"`
	Description   string     `gorm:"column:description;type:text"`
	Address       string     `gorm:"column:address;type:text"`
	City          string     `gorm:"column:city;type:varchar(100)"`
	Country       string     `gorm:"column:country;type:varchar(100)"`
	KYCStatus     string     `gorm:"column:kyc_status;type:varchar(50);default:'pending'"`
	IsActive      bool       `gorm:"column:is_active;default:true;index"`
	CreatedBy     string     `gorm:"column:created_by;type:uuid;index"`
	CreatedAt     time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;index"`
}

func (AccountModel) TableName() string {
	return "accounts"
}

func (m *AccountModel) ToDomain() *accountdomain.Account {
	if m == nil {
		return nil
	}
	return &accountdomain.Account{
		ID:            m.ID,
		Name:          m.Name,
		DisplayName:   m.DisplayName,
		Slug:          m.Slug,
		Email:         m.Email,
		Phone:         m.Phone,
		AccountTypeID: m.AccountTypeID,
		Status:        m.Status,
		LogoURL:       m.LogoURL,
		Website:       m.Website,
		Description:   m.Description,
		Address:       m.Address,
		City:          m.City,
		Country:       m.Country,
		KYCStatus:     m.KYCStatus,
		IsActive:      m.IsActive,
		CreatedBy:     m.CreatedBy,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		DeletedAt:     m.DeletedAt,
	}
}

func (m *AccountModel) FromDomain(account *accountdomain.Account) {
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
	m.Status = account.Status
	m.LogoURL = account.LogoURL
	m.Website = account.Website
	m.Description = account.Description
	m.Address = account.Address
	m.City = account.City
	m.Country = account.Country
	m.KYCStatus = account.KYCStatus
	m.IsActive = account.IsActive
	m.CreatedBy = account.CreatedBy
	m.CreatedAt = account.CreatedAt
	m.UpdatedAt = account.UpdatedAt
	m.DeletedAt = account.DeletedAt
}

// ============================================================
// ACCOUNT MEMBER MODEL
// ============================================================

type AccountMemberModel struct {
	ID        string     `gorm:"primaryKey;default:gen_random_uuid()"`
	AccountID string     `gorm:"column:account_id;type:uuid;not null;index"`
	UserID    string     `gorm:"column:user_id;type:uuid;not null;index"`
	Role      string     `gorm:"column:role;type:varchar(50);not null"`
	IsActive  bool       `gorm:"column:is_active;default:true"`
	InvitedBy string     `gorm:"column:invited_by;type:uuid;index"`
	JoinedAt  time.Time  `gorm:"column:joined_at;default:CURRENT_TIMESTAMP"`
	CreatedAt time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index"`
}

func (AccountMemberModel) TableName() string {
	return "account_members"
}

func (m *AccountMemberModel) ToDomain() *accountdomain.AccountMember {
	if m == nil {
		return nil
	}
	return &accountdomain.AccountMember{
		ID:        m.ID,
		AccountID: m.AccountID,
		UserID:    m.UserID,
		Role:      m.Role,
		IsActive:  m.IsActive,
		InvitedBy: m.InvitedBy,
		JoinedAt:  m.JoinedAt,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
	}
}

func (m *AccountMemberModel) FromDomain(member *accountdomain.AccountMember) {
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
// USER MODEL
// ============================================================
//
// UserModel is the persistence representation of a user.
//
// In this system, users are owned by the account module. The
// UserInfo projection (accountdomain.UserInfo) is derived from this
// model and exposed to other modules (events, audit logs, etc.).
//
// NOTE: field names may not exactly match your users table. Adjust
// the gorm column tags to match the actual schema.

type UserModel struct {
	ID          string     `gorm:"primaryKey;type:uuid"`
	Name        string     `gorm:"column:name;type:varchar(255);not null"`
	DisplayName string     `gorm:"column:display_name;type:varchar(255)"`
	Email       string     `gorm:"column:email;type:varchar(255);uniqueIndex;not null"`
	Phone       string     `gorm:"column:phone;type:varchar(50)"`
	AvatarURL   string     `gorm:"column:avatar_url;type:varchar(500)"`
	Bio         string     `gorm:"column:bio;type:text"`
	Location    string     `gorm:"column:location;type:varchar(255)"`
	Website     string     `gorm:"column:website;type:varchar(255)"`
	Slug        string     `gorm:"column:slug;type:varchar(50);uniqueIndex;not null"`
	SocialLinks types.JSONB      `gorm:"column:social_links;type:jsonb;default:'{}'"`
	IsActive    bool       `gorm:"column:is_active;default:true;index"`
	CreatedAt   time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;index"`
}

func (UserModel) TableName() string { return "users" }

func (m *UserModel) ToDomain() *accountdomain.User {
	if m == nil {
		return nil
	}
	return &accountdomain.User{
		ID:          m.ID,
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Email:       m.Email,
		Phone:       m.Phone,
		AvatarURL:   m.AvatarURL,
		Bio:         m.Bio,
		Location:    m.Location,
		Website:     m.Website,
		Slug:        m.Slug,
		SocialLinks: jsonbToStringMap(m.SocialLinks),
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
	}
}

func (m *UserModel) FromDomain(u *accountdomain.User) {
	if u == nil {
		return
	}
	m.ID = u.ID
	m.Name = u.Name
	m.DisplayName = u.DisplayName
	m.Email = u.Email
	m.Phone = u.Phone
	m.AvatarURL = u.AvatarURL
	m.Bio = u.Bio
	m.Location = u.Location
	m.Website = u.Website
	m.Slug = u.Slug
	m.SocialLinks = stringMapToJSONB(u.SocialLinks)
	m.IsActive = u.IsActive
	m.CreatedAt = u.CreatedAt
	m.UpdatedAt = u.UpdatedAt
	m.DeletedAt = u.DeletedAt
}

func (m *UserModel) ToUserInfo() *accountdomain.UserInfo {
	if m == nil {
		return nil
	}
	return &accountdomain.UserInfo{
		ID:          m.ID,
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Email:       m.Email,
		Phone:       m.Phone,
		AvatarURL:   m.AvatarURL,
		Bio:         m.Bio,
		Location:    m.Location,
		Website:     m.Website,
		SocialLinks: jsonbToStringMap(m.SocialLinks),
	}
}

// jsonbToStringMap converts a JSONB map to map[string]string, dropping non-string values.
func jsonbToStringMap(j types.JSONB) map[string]string {
	if j == nil {
		return map[string]string{}
	}
	out := make(map[string]string)
	for k, v := range j {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

// stringMapToJSONB converts a map[string]string to JSONB.
func stringMapToJSONB(m map[string]string) types.JSONB {
	if len(m) == 0 {
		return types.JSONB{}
	}
	out := make(types.JSONB, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}