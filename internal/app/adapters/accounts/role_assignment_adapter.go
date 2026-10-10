// internal/app/adapters/accounts/role_assignment_adapter.go

package accounts

import (
	"context"
	"fmt"

	accountdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
)

// RoleAssignmentAdapter implements accountdomain.RoleAssignment using
// the auth module's Casbin enforcer.
//
// It is the only place the account module touches Casbin. Everything
// else sees the port.
type RoleAssignmentAdapter struct {
	enforcer *authorization.Enforcer
}

func NewRoleAssignmentAdapter(enforcer *authorization.Enforcer) accountdomain.RoleAssignment {
	return &RoleAssignmentAdapter{enforcer: enforcer}
}

// Assign grants a role to a user in an account.
// Writes: ptype=g, v0=userID, v1=role, v2=account:<accountID>
func (a *RoleAssignmentAdapter) Assign(
	ctx context.Context,
	accountID, userID, role string,
) error {
	if accountID == "" || userID == "" || role == "" {
		return fmt.Errorf("accountID, userID, and role are required")
	}
	domain := accountdomain.AccountDomain(accountID)
	if domain == "" {
		return fmt.Errorf("invalid account ID: %q", accountID)
	}
	if _, err := a.enforcer.AddGroupingPolicy(userID, role, domain); err != nil {
		return fmt.Errorf("assign role: %w", err)
	}
	return nil
}

// RevokeAll removes every grouping rule for the (user, account) pair,
// regardless of role. Idempotent — safe when no rules exist.
func (a *RoleAssignmentAdapter) RevokeAll(
	ctx context.Context,
	accountID, userID string,
) error {
	if accountID == "" || userID == "" {
		return fmt.Errorf("accountID and userID are required")
	}
	domain := accountdomain.AccountDomain(accountID)
	if domain == "" {
		return fmt.Errorf("invalid account ID: %q", accountID)
	}
	// ptype=g, v0=userID, v1=*, v2=domain
	if _, err := a.enforcer.RemoveFilteredGroupingPolicy(0, userID, "", domain); err != nil {
		return fmt.Errorf("revoke all: %w", err)
	}
	return nil
}

// RevokeAllForAccount removes every grouping rule whose domain is the
// given account. Used when the account itself is deleted.
func (a *RoleAssignmentAdapter) RevokeAllForAccount(
	ctx context.Context,
	accountID string,
) error {
	if accountID == "" {
		return fmt.Errorf("accountID is required")
	}
	domain := accountdomain.AccountDomain(accountID)
	if domain == "" {
		return fmt.Errorf("invalid account ID: %q", accountID)
	}
	// ptype=g, v2=domain
	if _, err := a.enforcer.RemoveFilteredGroupingPolicy(2, domain); err != nil {
		return fmt.Errorf("revoke all for account: %w", err)
	}
	return nil
}

var _ accountdomain.RoleAssignment = (*RoleAssignmentAdapter)(nil)