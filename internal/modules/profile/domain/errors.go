// internal/modules/profile/domain/errors.go

package domain

import "errors"

// ============================================================
// CORE ERRORS
// ============================================================

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInstitutionNotFound = errors.New("institution not found")
	ErrInvalidUserID       = errors.New("invalid user ID")
	ErrInvalidInstitutionID = errors.New("invalid institution ID")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrInvalidScope        = errors.New("invalid scope")
)

// ============================================================
// TEAM MEMBER ERRORS
// ============================================================

var (
	ErrTeamMemberNotFound      = errors.New("team member not found")
	ErrTeamMemberAlreadyExists = errors.New("user is already a member of this team")
	ErrInvalidTeamType         = errors.New("invalid team type")
	ErrCannotRemoveSelf        = errors.New("cannot remove yourself from the team")
	ErrCannotRemoveOwner       = errors.New("cannot remove the team owner")
	ErrUserNotInTeam           = errors.New("user is not a member of this team")
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
)

// ============================================================
// MEMBER ROLE ERRORS
// ============================================================

var (
	ErrInvalidMemberRole   = errors.New("invalid member role")
	ErrCannotChangeOwnRole = errors.New("cannot change your own role")
)

// ============================================================
// DEPRECATED / ALIAS ERRORS (for backward compatibility)
// ============================================================

// Deprecated: Use ErrTeamInvitationNotFound instead
var ErrTeamInvitationNotFound = ErrInvitationNotFound

// Deprecated: Use ErrTeamInvitationAlreadyProcessed instead
var ErrTeamInvitationAlreadyProcessed = errors.New("invitation already processed")

// Deprecated: Use ErrTeamInvitationExpired instead
var ErrTeamInvitationExpired = ErrInvitationExpired

// Deprecated: Use ErrTeamInvitationEmailMismatch instead
var ErrTeamInvitationEmailMismatch = ErrInvitationEmailMismatch