package registrationdomain

import "context"

// PermissionChecker is the registration module's outbound port for
// resolving which accounts a user belongs to, and for asking whether
// the user holds a specific permission in a specific account.
//
// The port is intentionally narrow: only the two methods the
// registration service actually needs. Keeping it small makes it
// easy to fake in tests and prevents the registration module from
// growing a dependency on the full authorization surface.
type PermissionChecker interface {
	// GetUserAccountIDs returns every account UUID the user is a
	// member of. Returns an empty slice (not an error) for users who
	// belong to no accounts.
	GetUserAccountIDs(ctx context.Context, userID string) ([]string, error)

	// HasPermission reports whether userID holds the given
	// (resource, action) grant in the given domain.
	//
	// domain is expected to be a concrete account domain of the form
	// "account:<uuid>". Callers are responsible for constructing it.
	//
	// Used by the registration service to distinguish account-wide
	// reads (admin, trainer → registration:read) from own-scoped reads
	// (learner → registration:read_own) so the query can be filtered
	// accordingly.
	HasPermission(
		ctx context.Context,
		userID, domain, resource, action string,
	) (bool, error)
}