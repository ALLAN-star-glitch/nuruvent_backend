// internal/modules/team/service/service.go

package service

import (
    "context"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// Service defines the team module's business logic interface
type Service interface {
    // TEAM OPERATIONS
    CreatePersonalTeam(ctx context.Context, userID, userName string) (*teamdomain.Team, error)
    CreateInstitutionTeam(ctx context.Context, accountID, name, displayName, slug, createdBy string) (*teamdomain.Team, error)
    GetTeamByID(ctx context.Context, id string) (*teamdomain.Team, error)
    GetUserTeams(ctx context.Context, userID string) ([]*teamdomain.Team, error)
    UpdateTeam(ctx context.Context, id string, updates map[string]interface{}) (*teamdomain.Team, error)
    DeleteTeam(ctx context.Context, id string) error

    // MEMBER OPERATIONS
    AddMember(ctx context.Context, teamID, userID string, role teamdomain.MemberRole, addedBy string) (*teamdomain.Member, error)
    RemoveMember(ctx context.Context, teamID, userID, removedBy string) error
    UpdateMemberRole(ctx context.Context, teamID, userID string, newRole teamdomain.MemberRole, updatedBy string) (*teamdomain.Member, error)
    GetTeamMembers(ctx context.Context, teamID string, filters teamdomain.ListMembersFilters) ([]*teamdomain.Member, int64, error)
    GetUserMemberships(ctx context.Context, userID string) ([]*teamdomain.Member, error)
    LeaveTeam(ctx context.Context, teamID, userID string) error

    // INVITATION OPERATIONS
    InviteMember(ctx context.Context, cmd InviteMemberCommand) (*teamdomain.Invitation, error)
    ValidateInvitationToken(ctx context.Context, token string) (*teamdomain.Invitation, error)
    AcceptInvitation(ctx context.Context, token, userID string) (*teamdomain.Member, error)  // ✅ Changed: removed token returns
    DeclineInvitation(ctx context.Context, token, userID string) error
    ResendInvitation(ctx context.Context, invitationID string) (*teamdomain.Invitation, error)
    GetTeamInvitations(ctx context.Context, teamID string, filters teamdomain.ListInvitationsFilters) ([]*teamdomain.Invitation, int64, error)

    // ❌ REMOVED: RegisterAndAcceptInvitation - User registration belongs in Auth module, not Team

    // TEAM MEMBERSHIP QUERIES (For Auth Module)
    GetPersonalTeamByUserID(ctx context.Context, userID string) (*teamdomain.Team, error)
    GetInstitutionTeamByInstitutionID(ctx context.Context, institutionID string) (*teamdomain.Team, error)
    GetAccountByTeamID(ctx context.Context, teamID string) (*AccountInfo, error)
    GetUserTeamMemberships(ctx context.Context, userID string) ([]*TeamMemberInfo, error)
    GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error)
    GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error)
}

// InviteMemberCommand represents a request to invite a member
type InviteMemberCommand struct {
    TeamID    string
    Email     string
    Role      string
    InvitedBy string
}

// ❌ REMOVED: RegisterAndAcceptInvitationCommand - User registration belongs in Auth module

type AccountInfo struct {
    ID      string
    Name    string
    Email   string
    Phone   string
    Website string
}

type TeamMemberInfo struct {
    ID       string
    TeamID   string
    UserID   string
    Role     string
    IsActive bool
}

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

type TeamInfo struct {
    ID          string
    AccountID   string
    Name        string
    DisplayName string
    Slug        string
    Type        string
    MemberCount int
    Role        string
    CreatedAt   string
}

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