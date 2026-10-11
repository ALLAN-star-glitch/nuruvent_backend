package registration

import (
	"context"

	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	registrationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// PermissionCheckerAdapter bridges auth's PermissionChecker to the
// registration module's port.
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

// HasPermission reports whether userID holds the given (resource,
// action) grant in the given domain.
//
// The auth domain's PermissionChecker has this method; the adapter
// exposes it to the registration module under the same name so the
// semantics are unambiguous.
func (a *PermissionCheckerAdapter) HasPermission(
	ctx context.Context,
	userID, domain, resource, action string,
) (bool, error) {
	return a.checker.HasPermission(ctx, userID, domain, resource, action)
}

// compile-time assertion
var _ registrationDomain.PermissionChecker = (*PermissionCheckerAdapter)(nil)