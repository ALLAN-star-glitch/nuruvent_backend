package service
// internal/modules/team/service/service.go

import (
    "context"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// ============================================================
// TEAM SERVICE INTERFACE
// ============================================================

// Service defines the team module's business logic interface
type Service interface {
    // ============================================================
    // TEAM OPERATIONS
    // ============================================================

    CreatePersonalTeam(ctx context.Context, userID, userName string) (*teamdomain.Team, error)
    CreateInstitutionTeam(ctx context.Context, name, displayName, slug string) (*teamdomain.Team, error)
    GetTeamByID(ctx context.Context, id string) (*teamdomain.Team, error)
    GetUserTeams(ctx context.Context, userID string) ([]*teamdomain.Team, error)
    UpdateTeam(ctx context.Context, id string, updates map[string]interface{}) (*teamdomain.Team, error)
    DeleteTeam(ctx context.Context, id string) error

    // ============================================================
    // MEMBER OPERATIONS
    // ============================================================

    AddMember(ctx context.Context, teamID, userID string, role teamdomain.MemberRole, addedBy string) (*teamdomain.Member, error)
    RemoveMember(ctx context.Context, teamID, userID, removedBy string) error
    UpdateMemberRole(ctx context.Context, teamID, userID string, newRole teamdomain.MemberRole, updatedBy string) (*teamdomain.Member, error)
    GetTeamMembers(ctx context.Context, teamID string, filters teamdomain.ListMembersFilters) ([]*teamdomain.Member, int64, error)
    GetUserMemberships(ctx context.Context, userID string) ([]*teamdomain.Member, error)
    LeaveTeam(ctx context.Context, teamID, userID string) error

    // ============================================================
    // INVITATION OPERATIONS
    // ============================================================

    InviteMember(ctx context.Context, cmd InviteMemberCommand) (*teamdomain.Invitation, error)
    ValidateInvitationToken(ctx context.Context, token string) (*teamdomain.Invitation, error)
    AcceptInvitation(ctx context.Context, token, userID string) (*teamdomain.Member, string, string, error)
    RegisterAndAcceptInvitation(ctx context.Context, cmd RegisterAndAcceptInvitationCommand) (*teamdomain.Member, string, string, error)
    DeclineInvitation(ctx context.Context, token, userID string) error
    ResendInvitation(ctx context.Context, invitationID string) (*teamdomain.Invitation, error)
    GetTeamInvitations(ctx context.Context, teamID string, filters teamdomain.ListInvitationsFilters) ([]*teamdomain.Invitation, int64, error)
}

// ============================================================
// COMMANDS & DTOS
// ============================================================

// InviteMemberCommand represents a request to invite a member
type InviteMemberCommand struct {
    TeamID    string
    Email     string
    Role      teamdomain.MemberRole
    InvitedBy string
}

type RegisterAndAcceptInvitationCommand struct {
    Token    string
    Name     string
    Password string
    Phone    string
}


// MemberInfo represents team member information for API responses
type MemberInfo struct {
    ID          string
    TeamID      string
    UserID      string
    UserName    string
    UserEmail   string
    UserAvatar  string
    Role        string
    RoleDisplay string
    JoinedAt    string
    IsActive    bool
}

// TeamInfo represents team information for API responses
type TeamInfo struct {
    ID          string
    Name        string
    DisplayName string
    Slug        string
    Type        string
    MemberCount int
    Role        string
    CreatedAt   string
}

// InvitationInfo represents invitation information for API responses
type InvitationInfo struct {
    ID            string
    Email         string
    Role          string
    RoleDisplay   string
    Status        string
    ExpiresAt     string
    CreatedAt     string
    InvitedBy     string
    InvitedByName string
}