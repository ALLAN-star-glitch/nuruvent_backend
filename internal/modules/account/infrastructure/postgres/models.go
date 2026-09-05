// internal/modules/account/infrastructure/postgres/models.go

package postgres

import (
    "time"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
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
    m.CreatedBy = account.CreatedBy
    m.CreatedAt = account.CreatedAt
    m.UpdatedAt = account.UpdatedAt
    m.DeletedAt = account.DeletedAt
}

// ============================================================
// ACCOUNT MEMBER MODEL
// ============================================================

type AccountMemberModel struct {
    ID          string     `gorm:"primaryKey;default:gen_random_uuid()"`
    AccountID   string     `gorm:"column:account_id;type:uuid;not null;index"`
    UserID      string     `gorm:"column:user_id;type:uuid;not null;index"`
    Role        string     `gorm:"column:role;type:varchar(50);not null"`
    IsActive    bool       `gorm:"column:is_active;default:true"`
    InvitedBy   string     `gorm:"column:invited_by;type:uuid;index"`
    JoinedAt    time.Time  `gorm:"column:joined_at;default:CURRENT_TIMESTAMP"`
    CreatedAt   time.Time  `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
    UpdatedAt   time.Time  `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
    DeletedAt   *time.Time `gorm:"column:deleted_at;index"`
}

func (AccountMemberModel) TableName() string {
    return "account_members"
}

func (m *AccountMemberModel) ToDomain() *accountdomain.AccountMember {
    if m == nil {
        return nil
    }
    return &accountdomain.AccountMember{
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