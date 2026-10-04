package registrationdomain

import "context"

// PermissionChecker is the registration module's outbound port for
// resolving which accounts a user belongs to.
//
// We deliberately expose only GetUserAccountIDs rather than the full
// authorization surface — this keeps the port narrow and the
// dependency easy to fake in tests.
type PermissionChecker interface {
	// GetUserAccountIDs returns every account UUID the user is a
	// member of. Returns an empty slice (not an error) for users who
	// belong to no accounts.
	GetUserAccountIDs(ctx context.Context, userID string) ([]string, error)
}