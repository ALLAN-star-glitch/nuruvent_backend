// internal/modules/profile/domain/errors.go

package domain

import "errors"

var (
	// User errors
	ErrInvalidUserID      = errors.New("invalid user ID")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidUserEmail   = errors.New("invalid user email")
	
	// Account errors (replaces Institution errors)
	ErrInvalidAccountID      = errors.New("invalid account ID")
	ErrAccountNotFound       = errors.New("account not found")
	ErrAccountAlreadyExists  = errors.New("account already exists")
	ErrInvalidAccountType    = errors.New("invalid account type")
	ErrInvalidAccountStatus  = errors.New("invalid account status")
	
	// Deprecated: Use ErrInvalidAccountID and ErrAccountNotFound
	ErrInvalidInstitutionID = errors.New("invalid institution ID")
	ErrInstitutionNotFound  = errors.New("institution not found")
	
	// Permission errors
	ErrPermissionDenied = errors.New("permission denied")
	ErrUnauthorized     = errors.New("unauthorized")
	
	// Team errors
	ErrInvalidTeamID        = errors.New("invalid team ID")
	ErrTeamNotFound         = errors.New("team not found")
	ErrTeamAlreadyExists    = errors.New("team already exists")
	ErrTeamNameRequired     = errors.New("team name is required")
	
	// Team member errors
	ErrTeamMemberNotFound      = errors.New("team member not found")
	ErrTeamMemberAlreadyExists = errors.New("team member already exists")
	ErrNotTeamMember           = errors.New("user is not a member of this team")
	
	// Account member errors
	ErrAccountMemberNotFound      = errors.New("account member not found")
	ErrAccountMemberAlreadyExists = errors.New("account member already exists")
	ErrNotAccountMember           = errors.New("user is not a member of this account")
	ErrInvalidRole                = errors.New("invalid role")
	
	// Invalid scope errors (deprecated)
	ErrInvalidScope = errors.New("invalid scope")
	
	// Invalid organizer errors
	ErrInvalidOrganizerID   = errors.New("invalid organizer ID")
	ErrInvalidOrganizerType = errors.New("invalid organizer type")
	
	// Media errors
	ErrInvalidFileType = errors.New("invalid file type")
	ErrFileTooLarge    = errors.New("file too large")
	ErrUploadFailed    = errors.New("file upload failed")
	ErrNoFileProvided  = errors.New("no file provided")
)