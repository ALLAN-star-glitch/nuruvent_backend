// internal/modules/team/service/service.go

package service

import (
    "context"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// Service defines the team module's business logic interface
type Service interface {
    // TEAM OPERATIONS
    CreatePersonalTeam(ctx context.Context, userID, userName, role string) (*teamdomain.Team, error)
    CreateInstitutionTeam(ctx context.Context, accountID, name, displayName, slug, createdBy, role string) (*teamdomain.Team, error)
    GetTeamByID(ctx context.Context, id string) (*teamdomain.Team, error)
    GetUserTeams(ctx context.Context, userID string) ([]*teamdomain.Team, error)
    UpdateTeam(ctx context.Context, id string, updates map[string]interface{}) (*teamdomain.Team, error)
    DeleteTeam(ctx context.Context, id string) error

    // MEMBER OPERATIONS
    // ✅ REMOVED: role parameter - roles are inherited from account
    AddMember(ctx context.Context, teamID, userID, addedBy string) (*teamdomain.Member, error)
    // ✅ REMOVED: UpdateMemberRole - roles cannot be updated at team level
    RemoveMember(ctx context.Context, teamID, userID, removedBy string) error
    GetTeamMembers(ctx context.Context, teamID string, filters teamdomain.ListMembersFilters) ([]*teamdomain.Member, int64, error)
    GetUserMemberships(ctx context.Context, userID string) ([]*teamdomain.Member, error)
    LeaveTeam(ctx context.Context, teamID, userID string) error

    // INVITATION OPERATIONS
    // ✅ REMOVED: role from InviteMemberCommand - roles are inherited from account
    InviteMember(ctx context.Context, cmd InviteMemberCommand) (*teamdomain.Invitation, error)
    ValidateInvitationToken(ctx context.Context, token string) (*teamdomain.Invitation, error)
    AcceptInvitation(ctx context.Context, token, userID string) (*teamdomain.Member, error)
    DeclineInvitation(ctx context.Context, token, userID string) error
    ResendInvitation(ctx context.Context, invitationID string) (*teamdomain.Invitation, error)
    GetTeamInvitations(ctx context.Context, teamID string, filters teamdomain.ListInvitationsFilters) ([]*teamdomain.Invitation, int64, error)

    // TEAM MEMBERSHIP QUERIES (For Auth Module)
    GetPersonalTeamByUserID(ctx context.Context, userID string) (*teamdomain.Team, error)
    GetInstitutionTeamByInstitutionID(ctx context.Context, institutionID string) (*teamdomain.Team, error)
    GetAccountByTeamID(ctx context.Context, teamID string) (*AccountInfo, error)
    GetUserTeamMemberships(ctx context.Context, userID string) ([]*TeamMemberInfo, error)
    GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error)
    GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error)
}

// InviteMemberCommand represents a request to invite a member
// ✅ REMOVED: Role field - roles are inherited from account
type InviteMemberCommand struct {
    TeamID    string
    Email     string
    InvitedBy string
    // ❌ REMOVED: Role      string
}

// AccountInfo represents account information for a team
type AccountInfo struct {
    ID      string
    Name    string
    Email   string
    Phone   string
    Website string
}

// TeamMemberInfo represents team membership information
type TeamMemberInfo struct {
    ID       string
    TeamID   string
    UserID   string
    IsActive bool
    // ❌ REMOVED: Role     string
}

// MemberInfo represents member information for display
type MemberInfo struct {
    ID          string
    TeamID      string
    UserID      string
    UserName    string
    UserEmail   string
    UserAvatar  string
    JoinedAt    string
    IsActive    bool
    // ❌ REMOVED: Role        string
    // ❌ REMOVED: RoleDisplay string
}

// TeamInfo represents team information for display
type TeamInfo struct {
    ID          string
    AccountID   string
    Name        string
    DisplayName string
    Slug        string
    Type        string
    MemberCount int
    CreatedAt   string
    // ❌ REMOVED: Role        string
}

// InvitationInfo represents invitation information for display
type InvitationInfo struct {
    ID            string
    Email         string
    Status        string
    ExpiresAt     string
    CreatedAt     string
    InvitedBy     string
    InvitedByName string
    // ❌ REMOVED: Role          string
    // ❌ REMOVED: RoleDisplay   string
}