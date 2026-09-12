// internal/modules/team/teamdomain/errors.go

package teamdomain

import "errors"

// ============================================================
// TEAM ERRORS
// ============================================================

var (
    ErrTeamNotFound         = errors.New("team not found")
    ErrTeamAlreadyExists    = errors.New("team already exists")
    ErrInvalidTeamType      = errors.New("invalid team type")
    ErrTeamInactive         = errors.New("team is inactive")
    ErrTeamNameRequired     = errors.New("team name is required")
    ErrTeamSlugRequired     = errors.New("team slug is required")
)

// ============================================================
// MEMBER ERRORS
// ============================================================

var (
    ErrMemberNotFound       = errors.New("member not found")
    ErrMemberAlreadyExists  = errors.New("member already exists in this team")
    ErrInvalidRole          = errors.New("invalid role")
    ErrCannotRemoveSelf     = errors.New("cannot remove yourself from the team")
    ErrCannotRemoveOwner    = errors.New("cannot remove the team owner")
    ErrCannotChangeOwnRole  = errors.New("cannot change your own role")
    ErrUserNotInTeam        = errors.New("user is not a member of this team")
    ErrLastAdminCannotLeave = errors.New("cannot leave as the last admin of the team")
)

// ============================================================
// INVITATION ERRORS
// ============================================================

var (
    ErrInvitationNotFound        = errors.New("invitation not found")
    ErrInvitationExpired         = errors.New("invitation has expired")
    ErrInvitationAlreadyAccepted = errors.New("invitation already accepted")
    ErrInvitationAlreadyDeclined = errors.New("invitation already declined")
    ErrInvalidInvitationStatus   = errors.New("invalid invitation status")
    ErrInvalidInvitationToken    = errors.New("invalid invitation token")
    ErrInvitationEmailMismatch   = errors.New("invitation email does not match user")
    ErrInvitationCannotRegister  = errors.New("this invitation has expired. Please contact the admin for a new invitation")
)

// ============================================================
// PERMISSION ERRORS
// ============================================================

var (
    ErrPermissionDenied = errors.New("permission denied")
    ErrInsufficientRole = errors.New("insufficient role")
)


var ErrInvitationPending = errors.New("an invitation is already pending for this email")