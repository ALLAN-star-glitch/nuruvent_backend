// internal/modules/account/accountdomain/permission_checker.go

package accountdomain

import "context"

// PermissionChecker defines the permission-checking operations needed by
// the account module.
//
// The `domain` parameter uses the same string format as authdomain's
// PermissionChecker — see internal/shared/domains for the canonical
// builders:
//
//	"institution:team:{team_id}"
//	"personal:team:{team_id}"
//	"account:{account_id}"
//
// The account module constructs these via its own accountdomain facade
// (accountdomain.BuildTeamDomain, accountdomain.AccountDomain, etc.).
type PermissionChecker interface {
	// HasPermission checks if a user has a specific permission in a domain.
	HasPermission(ctx context.Context, userID, domain, resource, action string) (bool, error)

	// HasAnyPermission checks if a user has any of the given permissions in a domain.
	HasAnyPermission(ctx context.Context, userID, domain, resource string, actions ...string) (bool, error)

	// HasAllPermissions checks if a user has all of the given permissions in a domain.
	HasAllPermissions(ctx context.Context, userID, domain, resource string, actions ...string) (bool, error)

	// CanManageAccount checks if user can manage an account.
	// Domain: "account:{account_id}"
	CanManageAccount(ctx context.Context, userID, domain string) (bool, error)

	// CanManageAccountMembers checks if user can manage account members.
	// Domain: "account:{account_id}"
	CanManageAccountMembers(ctx context.Context, userID, domain string) (bool, error)

	// CanViewAccount checks if user can view an account.
	// Domain: "account:{account_id}"
	CanViewAccount(ctx context.Context, userID, domain string) (bool, error)
}