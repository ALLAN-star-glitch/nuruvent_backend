// internal/modules/account/accountdomain/role_assignment.go

package accountdomain

import "context"

// RoleAssignment keeps the authorization system in sync when account
// membership changes.
//
// This is an outbound port. The account module does not know or care
// which engine enforces roles — it only knows that role changes need
// to be reflected somewhere.
//
// The implementation lives in app/adapters/accounts and today
// delegates to the auth module's Casbin-backed role manager.
type RoleAssignment interface {
    // Assign grants a role to a user in an account.
    Assign(ctx context.Context, accountID, userID, role string) error

    // RevokeAll clears every role the user holds in the account.
    RevokeAll(ctx context.Context, accountID, userID string) error

    // RevokeAllForAccount clears every role for every user in the
    // account. Used when the account itself is deleted.
    RevokeAllForAccount(ctx context.Context, accountID string) error
}