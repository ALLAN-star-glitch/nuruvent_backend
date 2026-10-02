package attendance

import (
	"context"

	attendanceDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// PermissionCheckerAdapter bridges auth's PermissionChecker to the
// attendance module's port.
//
// The attendance module only needs to know which accounts a user
// belongs to, so the adapter exposes just that method.
type PermissionCheckerAdapter struct {
	checker authDomain.PermissionChecker
}

// NewPermissionCheckerAdapter constructs the adapter.
func NewPermissionCheckerAdapter(
	checker authDomain.PermissionChecker,
) attendanceDomain.PermissionChecker {
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
var _ attendanceDomain.PermissionChecker = (*PermissionCheckerAdapter)(nil)