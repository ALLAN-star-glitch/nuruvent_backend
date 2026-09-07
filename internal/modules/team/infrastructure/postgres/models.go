// internal/modules/team/infrastructure/postgres/models.go

package postgres

import (
    "time"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// ============================================================
// TEAM MODEL
// ============================================================

type TeamModel struct {
    ID          string     `gorm:"primaryKey;default:gen_random_uuid()"`
    AccountID   string     `gorm:"not null;index"`
    Name        string     `gorm:"not null"`
    DisplayName string     `gorm:"column:display_name"`
    Slug        string     `gorm:"uniqueIndex;not null"`
    Type        string     `gorm:"column:type;not null"`
    IsActive    bool       `gorm:"default:true"`
    CreatedAt   time.Time  `gorm:"column:created_at"`
    UpdatedAt   time.Time  `gorm:"column:updated_at"`
    DeletedAt   *time.Time `gorm:"column:deleted_at;index"`
}

func (TeamModel) TableName() string {
    return "teams"
}

func (m *TeamModel) ToDomain() *teamdomain.Team {
    if m == nil {
        return nil
    }
    return &teamdomain.Team{
        ID:          m.ID,
        AccountID:   m.AccountID,  // ✅ Add this
        Name:        m.Name,
        DisplayName: m.DisplayName,
        Slug:        m.Slug,
        Type:        teamdomain.TeamType(m.Type),
        IsActive:    m.IsActive,
        CreatedAt:   m.CreatedAt,
        UpdatedAt:   m.UpdatedAt,
        DeletedAt:   m.DeletedAt,
    }
}


func (m *TeamModel) FromDomain(team *teamdomain.Team) {
    if team == nil {
        return
    }
    m.ID = team.ID
    m.AccountID = team.AccountID  // ✅ Add this - THIS WAS THE MISSING LINE
    m.Name = team.Name
    m.DisplayName = team.DisplayName
    m.Slug = team.Slug
    m.Type = string(team.Type)
    m.IsActive = team.IsActive
    m.CreatedAt = team.CreatedAt
    m.UpdatedAt = team.UpdatedAt
    m.DeletedAt = team.DeletedAt
}

// ============================================================
// MEMBER MODEL
// ============================================================

type MemberModel struct {
    ID        string     `gorm:"primaryKey;default:gen_random_uuid()"`
    TeamID    string     `gorm:"column:team_id;not null;index"`
    UserID    string     `gorm:"column:user_id;not null;index"`
    Role      string     `gorm:"column:role;not null"`
    IsActive  bool       `gorm:"default:true"`
    JoinedAt  time.Time  `gorm:"column:joined_at"`
    CreatedAt time.Time  `gorm:"column:created_at"`
    UpdatedAt time.Time  `gorm:"column:updated_at"`
    DeletedAt *time.Time `gorm:"column:deleted_at;index"`
}

func (MemberModel) TableName() string {
    return "team_members"
}

func (m *MemberModel) ToDomain() *teamdomain.Member {
    if m == nil {
        return nil
    }
    return &teamdomain.Member{
        ID:        m.ID,
        TeamID:    m.TeamID,
        UserID:    m.UserID,
        Role:      teamdomain.MemberRole(m.Role),
        IsActive:  m.IsActive,
        JoinedAt:  m.JoinedAt,
        CreatedAt: m.CreatedAt,
        UpdatedAt: m.UpdatedAt,
        DeletedAt: m.DeletedAt,
    }
}

func (m *MemberModel) FromDomain(member *teamdomain.Member) {
    if member == nil {
        return
    }
    m.ID = member.ID
    m.TeamID = member.TeamID
    m.UserID = member.UserID
    m.Role = string(member.Role)
    m.IsActive = member.IsActive
    m.JoinedAt = member.JoinedAt
    m.CreatedAt = member.CreatedAt
    m.UpdatedAt = member.UpdatedAt
    m.DeletedAt = member.DeletedAt
}

// ============================================================
// INVITATION MODEL
// ============================================================

type InvitationModel struct {
    ID         string     `gorm:"primaryKey;default:gen_random_uuid()"`
    TeamID     string     `gorm:"column:team_id;not null;index"`
    Email      string     `gorm:"column:email;not null;index"`
    Role       string     `gorm:"column:role;not null"`
    Token      string     `gorm:"column:token;not null;uniqueIndex"`
    Status     string     `gorm:"column:status;not null;default:'pending'"`
    InvitedBy  string     `gorm:"column:invited_by;not null"`
    ExpiresAt  time.Time  `gorm:"column:expires_at;not null"`
    AcceptedAt *time.Time `gorm:"column:accepted_at"`
    DeclinedAt *time.Time `gorm:"column:declined_at"`
    CreatedAt  time.Time  `gorm:"column:created_at"`
    UpdatedAt  time.Time  `gorm:"column:updated_at"`
    DeletedAt  *time.Time `gorm:"column:deleted_at;index"`
}

func (InvitationModel) TableName() string {
    return "team_invitations"
}

func (m *InvitationModel) ToDomain() *teamdomain.Invitation {
    if m == nil {
        return nil
    }
    return &teamdomain.Invitation{
        ID:         m.ID,
        TeamID:     m.TeamID,
        Email:      m.Email,
        Role:       teamdomain.MemberRole(m.Role),
        Token:      m.Token,
        Status:     teamdomain.InvitationStatus(m.Status),
        InvitedBy:  m.InvitedBy,
        ExpiresAt:  m.ExpiresAt,
        AcceptedAt: m.AcceptedAt,
        DeclinedAt: m.DeclinedAt,
        CreatedAt:  m.CreatedAt,
        UpdatedAt:  m.UpdatedAt,
        DeletedAt:  m.DeletedAt,
    }
}

func (m *InvitationModel) FromDomain(invitation *teamdomain.Invitation) {
    if invitation == nil {
        return
    }
    m.ID = invitation.ID
    m.TeamID = invitation.TeamID
    m.Email = invitation.Email
    m.Role = string(invitation.Role)
    m.Token = invitation.Token
    m.Status = string(invitation.Status)
    m.InvitedBy = invitation.InvitedBy
    m.ExpiresAt = invitation.ExpiresAt
    m.AcceptedAt = invitation.AcceptedAt
    m.DeclinedAt = invitation.DeclinedAt
    m.CreatedAt = invitation.CreatedAt
    m.UpdatedAt = invitation.UpdatedAt
    m.DeletedAt = invitation.DeletedAt
}