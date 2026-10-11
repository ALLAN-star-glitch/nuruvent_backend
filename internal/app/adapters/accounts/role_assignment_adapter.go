// internal/app/adapters/accounts/role_assignment_adapter.go

package accounts

import (
	"context"
	"fmt"

	accountdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	authseeder "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/auth_seeder"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"gorm.io/gorm"
)

// roleAssignmentAdapter implements accountdomain.RoleAssignment on top
// of the auth module's Casbin enforcer.
//
// SQL (account_members) is the source of truth for a user's role in an
// account. Casbin grouping rules are a derived cache that must stay in
// sync. Assign / RevokeAll / RevokeAllForAccount write the cache
// directly; ReconcileAll repairs it by diffing against SQL.
type roleAssignmentAdapter struct {
	enforcer *authorization.Enforcer
	db       *gorm.DB
}

// NewRoleAssignmentAdapter bridges the auth module's Casbin enforcer
// and the shared DB handle to the account domain's RoleAssignment port.
//
// The DB handle is needed for ReconcileAll, which reads account_members
// and re-aligns Casbin grouping rules.
func NewRoleAssignmentAdapter(
	enforcer *authorization.Enforcer,
	db *gorm.DB,
) accountdomain.RoleAssignment {
	return &roleAssignmentAdapter{
		enforcer: enforcer,
		db:       db,
	}
}

// ============================================================
// ASSIGN
// ============================================================

// Assign grants a role to a user in an account.
//
// Casbin's AddGroupingPolicy is additive — calling it twice for the
// same (user, role, domain) is a no-op, but calling it with a
// *different* role leaves the old rule in place. Callers that change
// a role must RevokeAll first. UpdateMemberRole in the account
// service already does this.
func (a *roleAssignmentAdapter) Assign(
	ctx context.Context,
	accountID, userID, role string,
) error {
	domain := accountdomain.AccountDomain(accountID)
	if domain == "" {
		return fmt.Errorf("invalid account ID: %q", accountID)
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if role == "" {
		return fmt.Errorf("role is required")
	}

	_, err := a.enforcer.AddGroupingPolicy(userID, role, domain)
	if err != nil {
		return fmt.Errorf("add grouping policy: %w", err)
	}
	return nil
}

// ============================================================
// REVOKE ALL (per user, per account)
// ============================================================

// RevokeAll clears every role the user holds in the account.
func (a *roleAssignmentAdapter) RevokeAll(
	ctx context.Context,
	accountID, userID string,
) error {
	domain := accountdomain.AccountDomain(accountID)
	if domain == "" {
		return fmt.Errorf("invalid account ID: %q", accountID)
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	// ptype=g, v0=userID, v1=* (any role), v2=domain
	if _, err := a.enforcer.RemoveFilteredGroupingPolicy(0, userID, "", domain); err != nil {
		return fmt.Errorf("remove filtered grouping policy: %w", err)
	}
	return nil
}

// ============================================================
// REVOKE ALL FOR ACCOUNT
// ============================================================

// RevokeAllForAccount clears every role for every user in the account.
// Used when the account itself is deleted.
func (a *roleAssignmentAdapter) RevokeAllForAccount(
	ctx context.Context,
	accountID string,
) error {
	domain := accountdomain.AccountDomain(accountID)
	if domain == "" {
		return fmt.Errorf("invalid account ID: %q", accountID)
	}

	// ptype=g, v2=domain
	if _, err := a.enforcer.RemoveFilteredGroupingPolicy(2, domain); err != nil {
		return fmt.Errorf("remove filtered grouping policy by domain: %w", err)
	}
	return nil
}

// ============================================================
// RECONCILE ALL
// ============================================================

// ReconcileAll re-aligns Casbin grouping rules with account_members.
//
// It reads every active membership row, diffs it against the current
// grouping rules for concrete account domains, and applies the
// differences:
//
//   - SQL row without a matching rule   → rule is inserted.
//   - Rule with no matching active row  → rule is deleted.
//   - Role disagreement                 → rule is replaced.
//
// The platform domain, the role-hierarchy wildcard rule, and any rule
// whose subject is not a real user are left untouched. See
// authseeder.ReconcileAccountMembers for the full contract.
//
// Idempotent. Safe to run at startup and after every role mutation.
func (a *roleAssignmentAdapter) ReconcileAll(ctx context.Context) error {
	if a.db == nil {
		return fmt.Errorf("reconcile: db handle is nil")
	}
	if a.enforcer == nil {
		return fmt.Errorf("reconcile: enforcer is nil")
	}

	if _, err := authseeder.ReconcileAccountMembers(ctx, a.db, a.enforcer); err != nil {
		return fmt.Errorf("reconcile account members: %w", err)
	}
	return nil
}

// Ensure the adapter satisfies the port at compile time.
var _ accountdomain.RoleAssignment = (*roleAssignmentAdapter)(nil)