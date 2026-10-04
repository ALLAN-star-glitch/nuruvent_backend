package registration

import (
	"context"

	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	registrationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// PermissionCheckerAdapter bridges auth's PermissionChecker to the
// registration module's port.
//
// The registration module only needs to know which accounts a user
// belongs to, so the adapter exposes just that method.
type PermissionCheckerAdapter struct {
	checker authDomain.PermissionChecker
}

// NewPermissionCheckerAdapter constructs the adapter.
func NewPermissionCheckerAdapter(
	checker authDomain.PermissionChecker,
) registrationDomain.PermissionChecker {
	return &PermissionCheckerAdapter{checker: checker}
}

// GetUserAccountIDs returns every account UUID the user is a member
// of. Delegates directly to the auth domain's implementation.
func (a *PermissionCheckerAdapter) GetUserAccountIDs(
	ctx context.Context,
	userID string,
) ([]string, error) {
	return a.checker.GetUserAccountIDs(ctx, userID)
}

// compile-time assertion
var _ registrationDomain.PermissionChecker = (*PermissionCheckerAdapter)(nil)