// internal/modules/account/accountdomain/casbin_service.go

package accountdomain

import "context"

// CasbinService defines the Casbin operations the account module needs.
//
// This is an outbound port. The concrete implementation lives in
// app/adapters/accounts and wraps the auth module's enforcer.
//
// The account module uses this port to keep the DB and Casbin in sync when
// membership changes. It never talks to Casbin directly.
type CasbinService interface {
	// AssignAccountRole writes a grouping rule binding a user to a role in
	// the given account's domain.
	//
	//	ptype=g, v0=userID, v1=role, v2=account:<accountID>
	AssignAccountRole(ctx context.Context, accountID, userID, role string) error

	// RemoveAccountRole removes every grouping rule for the given user in
	// the given account's domain.
	//
	// Safe against role drift: it removes all rules for the (user, account)
	// pair regardless of role.
	RemoveAccountRole(ctx context.Context, accountID, userID string) error

	// RemoveAllAccountRoles removes every grouping rule whose domain is the
	// given account's domain. Used when an entire account is deleted.
	RemoveAllAccountRoles(ctx context.Context, accountID string) error
}