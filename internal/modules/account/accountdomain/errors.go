// internal/modules/account/accountdomain/errors.go

package accountdomain

import "errors"

// Account errors
var (
    ErrAccountNotFound         = errors.New("account not found")
    ErrAccountAlreadyExists    = errors.New("account already exists")
    ErrAccountTypeNotFound     = errors.New("account type not found")
    ErrAccountInvalidType      = errors.New("invalid account type")
    ErrAccountInactive         = errors.New("account is inactive")
    ErrAccountSuspended        = errors.New("account is suspended")
    ErrAccountEmailRequired    = errors.New("account email is required")
    ErrAccountNameRequired     = errors.New("account name is required")
    ErrAccountSlugRequired     = errors.New("account slug is required")
)

// Account member errors
var (
    ErrAccountMemberNotFound     = errors.New("account member not found")
    ErrAccountMemberAlreadyExists = errors.New("account member already exists")
    ErrInvalidRole               = errors.New("invalid role")
    ErrCannotRemoveSelf          = errors.New("cannot remove yourself from the account")
    ErrCannotChangeOwnRole       = errors.New("cannot change your own role")
    ErrLastAdminCannotLeave      = errors.New("cannot leave as the last admin of the account")
    ErrUserNotInAccount          = errors.New("user is not a member of this account")
)

// Permission errors
var (
    ErrPermissionDenied = errors.New("permission denied")
    ErrInsufficientRole = errors.New("insufficient role")



)


var ErrForbidden = errors.New("forbidden")