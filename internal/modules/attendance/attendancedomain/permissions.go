package attendancedomain

import "context"

// PermissionChecker exposes the subset of authorization queries the
// attendance module needs.
//
// This is a port — the concrete implementation lives in the auth
// module (or an adapter that wraps it) and is injected into the
// attendance service's Dependencies.
//
// We deliberately expose only GetUserAccountIDs rather than the full
// auth-domain interface: the attendance module resolves ownership
// through accounts (team → account → user), not through per-resource
// permission checks.
type PermissionChecker interface {
	// GetUserAccountIDs returns every account UUID the user is a
	// member of. The values are bare UUIDs, not "account:<uuid>"
	// domain strings.
	GetUserAccountIDs(ctx context.Context, userID string) ([]string, error)
}