package authdomain

import "errors"

// ============================================================
// AUTH DOMAIN ERRORS
// ============================================================

// User errors (formerly Account)
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserExists        = errors.New("user already exists")
	ErrUserInactive      = errors.New("user is inactive")
	ErrUserAlreadyActive = errors.New("user already active")
	ErrUserNotVerified   = errors.New("user email not verified")
)

// Account type errors
var (
	ErrInvalidAccountType      = errors.New("invalid account type")
	ErrAccountTypeNotFound     = errors.New("account type not found")
	ErrProfessionalTypeNotFound = errors.New("professional type not found")
)

// Credential errors
var (
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidPhone    = errors.New("invalid phone number")
	ErrWeakPassword    = errors.New("password too weak")
)

// Authentication errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account is locked")
	ErrTooManyAttempts    = errors.New("too many failed attempts")
)

// OTP errors
var (
	ErrInvalidOTP   = errors.New("invalid OTP")
	ErrExpiredOTP   = errors.New("OTP has expired")
	ErrOTPNotFound  = errors.New("OTP not found")
)

// Token errors
var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrRevokedToken  = errors.New("token has been revoked")
	ErrTokenNotFound = errors.New("token not found")
)

// Institution errors
var (
	ErrInstitutionNotFound = errors.New("institution not found")
	ErrInstitutionExists   = errors.New("institution already exists")
)

// Team member errors
var (
	ErrTeamMemberNotFound     = errors.New("team member not found")
	ErrTeamMemberExists       = errors.New("team member already exists")
	ErrInvalidRole            = errors.New("invalid role")
	ErrMemberAlreadyInTeam    = errors.New("member already in this team")
	ErrTeamMemberNotActive    = errors.New("team member is not active")
	ErrCannotRemoveOwner      = errors.New("cannot remove owner from team")
)

// Team type errors (NEW)
var (
	ErrTeamTypeNotFound     = errors.New("team type not found")
	ErrTeamTypeExists       = errors.New("team type already exists")
	ErrInvalidTeamType      = errors.New("invalid team type")
	ErrTeamTypeInUse        = errors.New("team type is in use and cannot be deleted")
	ErrInvalidTeamTypeForInstitution = errors.New("institution team must have an institution_id")
	ErrInvalidTeamTypeForPersonal    = errors.New("personal team must NOT have an institution_id")
)

// Invitation errors
var (
	ErrInvitationNotFound  = errors.New("invitation not found")
	ErrInvitationExpired   = errors.New("invitation has expired")
	ErrInvitationInvalid   = errors.New("invalid invitation")
	ErrInvitationAlreadyUsed = errors.New("invitation already used")
)

// Permission errors
var (
	ErrPermissionDenied    = errors.New("permission denied")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrInsufficientRole    = errors.New("insufficient role")
)





// ============================================================
// DEPRECATED: Keep for backward compatibility
// ============================================================

// Deprecated: Use ErrUserNotFound instead
var ErrAccountNotFound = ErrUserNotFound

// Deprecated: Use ErrUserExists instead
var ErrAccountExists = ErrUserExists

// Deprecated: Use ErrUserInactive instead
var ErrAccountInactive = ErrUserInactive

// Deprecated: Use ErrUserAlreadyActive instead
var ErrAccountAlreadyActive = ErrUserAlreadyActive

// Deprecated: Use ErrUserNotVerified instead
var ErrAccountNotVerified = ErrUserNotVerified


var (
   
    ErrInstitutionTypeNotFound = errors.New("institution type not found")
    ErrInvalidProfessionalType = errors.New("invalid professional type")
    ErrInvalidInstitutionType = errors.New("invalid institution type")
  


)