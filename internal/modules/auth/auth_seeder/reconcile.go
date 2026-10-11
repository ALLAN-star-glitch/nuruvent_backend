// internal/modules/auth/auth_seeder/reconcile.go

package authseeder

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"gorm.io/gorm"
)

// ============================================================
// ACCOUNT MEMBERSHIP RECONCILIATION
// ============================================================
//
// Source of truth: account_members.role (SQL).
// Derived cache:   casbin_rule grouping rules (g, user, role, account:<id>).
//
// If the two disagree, the user has powers the UI doesn't show — or
// shows powers they don't have. This function makes Casbin match SQL.
//
// Invariants enforced:
//
//   1. Every active account_members row has exactly one matching
//      grouping rule with the same role.
//   2. No grouping rule exists for a (user, account) pair that has no
//      active membership row.
//   3. No grouping rule exists whose role disagrees with the SQL role.
//   4. No grouping rule exists whose account domain has no real account.
//
// Idempotent. Safe to run at startup and after every role mutation.
//
// NOTE: This is NOT a seeder. It is a maintenance invariant. It runs
// from two places:
//
//   - Application startup (see cmd/api/main.go), once per boot.
//   - Account-member mutations (see the account service's
//     RoleAssignment.ReconcileAll hook), after every add / remove /
//     update-role / leave.
//
// It is deliberately not exposed through cmd/seed — the seeder
// framework assumes it controls the transaction, and this function
// writes to Casbin through a separate connection.

// AccountMemberRow mirrors the columns we need from account_members.
// Declared here rather than imported so the reconciliation is
// self-contained and can run before the Account module is wired.
type AccountMemberRow struct {
	UserID    string `gorm:"column:user_id"`
	AccountID string `gorm:"column:account_id"`
	Role      string `gorm:"column:role"`
	IsActive  bool   `gorm:"column:is_active"`
}

func (AccountMemberRow) TableName() string {
	return "account_members"
}

// membershipKey identifies a (user, account) pair. One key type for
// both the SQL side and the Casbin side of the diff.
type membershipKey struct {
	userID    string
	accountID string
}

// ReconcileResult reports what changed. Useful for tests and logging.
type ReconcileResult struct {
	Inserted int
	Updated  int
	Deleted  int
	Skipped  int
}

func (r ReconcileResult) Changed() bool {
	return r.Inserted > 0 || r.Updated > 0 || r.Deleted > 0
}

func (r ReconcileResult) String() string {
	return fmt.Sprintf(
		"inserted=%d updated=%d deleted=%d skipped=%d",
		r.Inserted, r.Updated, r.Deleted, r.Skipped,
	)
}

// ReconcileAccountMembers aligns Casbin grouping rules with the
// account_members table. See the package comment above for invariants.
func ReconcileAccountMembers(
	ctx context.Context,
	db *gorm.DB,
	enforcer *authorization.Enforcer,
) (ReconcileResult, error) {
	var result ReconcileResult

	// ------------------------------------------------------------
	// 1. Load the authoritative state from SQL.
	// ------------------------------------------------------------
	var members []AccountMemberRow
	if err := db.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&members).Error; err != nil {
		return result, fmt.Errorf("load account_members: %w", err)
	}

	want := make(map[membershipKey]string, len(members))
	for _, m := range members {
		if m.UserID == "" || m.AccountID == "" || m.Role == "" {
			log.Printf("[reconcile] skipping malformed membership row: user=%q account=%q role=%q",
				m.UserID, m.AccountID, m.Role)
			result.Skipped++
			continue
		}
		if !authdomain.IsValidAccountRole(m.Role) {
			log.Printf("[reconcile] skipping membership with unknown role: user=%q account=%q role=%q",
				m.UserID, m.AccountID, m.Role)
			result.Skipped++
			continue
		}
		want[membershipKey{m.UserID, m.AccountID}] = m.Role
	}

	// ------------------------------------------------------------
	// 2. Load the current state from Casbin.
	//
	// Only account-domain grouping rules are in scope. The platform
	// domain and the role-hierarchy wildcard rule are untouched.
	// ------------------------------------------------------------
	allGrouping, err := enforcer.GetGroupingPolicy()
	if err != nil {
		return result, fmt.Errorf("load grouping policies: %w", err)
	}

	have := make(map[membershipKey][]string)
	for _, rule := range allGrouping {
		if len(rule) < 3 {
			continue
		}
		userID, role, domain := rule[0], rule[1], rule[2]

		// Only reconcile concrete account domains.
		accountID := authdomain.ExtractAccountIDFromDomain(domain)
		if accountID == "" {
			// Platform domain, wildcard, or malformed — not ours.
			continue
		}

		// The role-hierarchy rule uses "*" as the subject, not a user.
		// Skip anything where the subject isn't a user ID or where the
		// role isn't a known account role.
		if userID == "*" || !authdomain.IsValidAccountRole(role) {
			continue
		}

		have[membershipKey{userID, accountID}] = rule
	}

	// ------------------------------------------------------------
	// 3. Diff and apply.
	// ------------------------------------------------------------

	// 3a. Deletes and updates: anything in Casbin that disagrees with SQL.
	var toDelete [][]string
	var toAdd [][]string

	for key, rule := range have {
		wantRole, ok := want[key]
		if !ok {
			// No active SQL row — the rule is stale.
			log.Printf("[reconcile] stale grouping rule: user=%s account=%s role=%s (no active membership)",
				key.userID, key.accountID, rule[1])
			toDelete = append(toDelete, rule)
			result.Deleted++
			continue
		}
		if rule[1] != wantRole {
			// Role drifted — replace.
			log.Printf("[reconcile] role drift: user=%s account=%s casbin=%s sql=%s",
				key.userID, key.accountID, rule[1], wantRole)
			toDelete = append(toDelete, rule)
			toAdd = append(toAdd, []string{
				key.userID, wantRole, authdomain.AccountDomain(key.accountID),
			})
			result.Updated++
			continue
		}
		// Agreement — leave it.
		result.Skipped++
	}

	// 3b. Inserts: SQL rows with no Casbin rule.
	for key, wantRole := range want {
		if _, ok := have[key]; ok {
			continue
		}
		log.Printf("[reconcile] missing grouping rule: user=%s account=%s role=%s",
			key.userID, key.accountID, wantRole)
		toAdd = append(toAdd, []string{
			key.userID, wantRole, authdomain.AccountDomain(key.accountID),
		})
		result.Inserted++
	}

	// 3c. Apply. Deletes first so a replace never collides with the old rule.
	if len(toDelete) > 0 {
		if _, err := enforcer.RemoveGroupingPolicies(toDelete); err != nil {
			return result, fmt.Errorf("remove stale grouping rules: %w", err)
		}
	}
	if len(toAdd) > 0 {
		if _, err := enforcer.AddGroupingPolicies(toAdd); err != nil {
			return result, fmt.Errorf("add missing grouping rules: %w", err)
		}
	}

	// ------------------------------------------------------------
	// 4. Reload the enforcer so the changes are visible immediately.
	// ------------------------------------------------------------
	if result.Changed() {
		if err := enforcer.LoadPolicy(); err != nil {
			return result, fmt.Errorf("reload policies after reconcile: %w", err)
		}
	}

	log.Printf("[reconcile] account memberships: %s", result.String())
	return result, nil
}

// ============================================================
// SERVICE HOOK
// ============================================================

// ReconcileMemberRoles is a thin wrapper the Account member service
// calls after add / update-role / remove / leave, via the
// RoleAssignment port.
func ReconcileMemberRoles(
	ctx context.Context,
	db *gorm.DB,
	enforcer *authorization.Enforcer,
) error {
	_, err := ReconcileAccountMembers(ctx, db, enforcer)
	return err
}