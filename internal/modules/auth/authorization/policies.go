// internal/modules/auth/authorization/policies.go

package authorization

import (
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// ============================================================
// STATIC POLICY SET
// ============================================================
//
// All policies are static. Domains use wildcards where the same grant
// applies to every account:
//
//   "platform"   - literal, one domain, Nuruvent staff only
//   "account:*"  - pattern, matches every "account:<uuid>" domain
//
// The Casbin matcher uses keyMatch2(r.dom, p.dom), so a wildcard pattern
// in the policy matches the concrete domain at check time.
//
// Teams are NOT authorization domains. Team membership is enforced as
// data (team_members) in the service layer, not as Casbin policies.
//
// ============================================================
// INVARIANTS
// ============================================================
//
//   1. Every policy's domain is either "platform" or "account:*".
//   2. No policy grants a role to a subject that isn't a known role.
//   3. No policy references a resource or action that isn't a known
//      constant.
//   4. "own"-scoped actions (_own) are granted only where the service
//      layer is prepared to apply an ownership filter.
//
// ValidatePolicies() enforces these at seed time.

// ============================================================
// POLICY VALIDATION
// ============================================================

// ValidatePolicies checks the invariant list above. Returns an error
// describing the first violation found. Called by the seeder before
// policies are written to the enforcer.
func ValidatePolicies() error {
	all := [][]string{}
	all = append(all, GetPlatformPolicies()...)
	all = append(all, GetAccountPolicies()...)

	validDomains := map[string]bool{
		authdomain.DomainPlatform: true,
		AccountDomainWildcard:     true,
	}

	for i, p := range all {
		if len(p) < 4 {
			return fmt.Errorf("policy %d: expected 4 fields, got %d", i, len(p))
		}
		subject, domain, resource, action := p[0], p[1], p[2], p[3]

		if !authdomain.IsValidRole(subject) {
			return fmt.Errorf("policy %d: unknown subject %q", i, subject)
		}
		if !validDomains[domain] {
			return fmt.Errorf("policy %d: invalid domain %q (must be %q or %q)",
				i, domain, authdomain.DomainPlatform, AccountDomainWildcard)
		}

		res := authdomain.Resource(resource)
		if !isKnownResource(res) {
			return fmt.Errorf("policy %d: unknown resource %q", i, resource)
		}
		act := authdomain.Action(action)
		if !isKnownAction(act) {
			return fmt.Errorf("policy %d: unknown action %q", i, action)
		}
	}
	return nil
}

func isKnownResource(r authdomain.Resource) bool {
	for _, known := range authdomain.GetAllResources() {
		if r == known {
			return true
		}
	}
	return false
}

func isKnownAction(a authdomain.Action) bool {
	for _, known := range authdomain.GetAllActions() {
		if a == known {
			return true
		}
	}
	return false
}

// ============================================================
// PLATFORM POLICIES
// ============================================================

// GetPlatformPolicies returns policies for the "platform" domain.
func GetPlatformPolicies() [][]string {
	superAdmin := authdomain.RoleSuperAdmin.String()
	admin := authdomain.RoleAdmin.String()
	guest := authdomain.RoleGuest.String()
	platform := authdomain.DomainPlatform

	var policies [][]string

	// ---- Super Admin: full platform management ----
	superAdminPolicies := [][]string{
		{superAdmin, platform, authdomain.ResourcePlatform.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourceUser.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourceInstitution.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourceAccount.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourceEvent.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourceCertificate.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourceAttendee.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourcePayment.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourcePayout.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourceMember.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourceProfile.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourceDashboard.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourceAnalytics.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourceNotification.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourceMedia.String(), authdomain.ActionManage.String()},
		{superAdmin, platform, authdomain.ResourceTeam.String(), authdomain.ActionManage.String()},
	}
	policies = append(policies, superAdminPolicies...)

	// ---- Admin: platform management, read-mostly ----
	adminPolicies := [][]string{
		{admin, platform, authdomain.ResourceUser.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourceUser.String(), authdomain.ActionUpdate.String()},
		{admin, platform, authdomain.ResourceUser.String(), authdomain.ActionDelete.String()},

		{admin, platform, authdomain.ResourceInstitution.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourceInstitution.String(), authdomain.ActionUpdate.String()},
		{admin, platform, authdomain.ResourceInstitution.String(), authdomain.ActionDelete.String()},

		{admin, platform, authdomain.ResourceAccount.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourceAccount.String(), authdomain.ActionUpdate.String()},
		{admin, platform, authdomain.ResourceAccount.String(), authdomain.ActionDelete.String()},

		{admin, platform, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{admin, platform, authdomain.ResourceEvent.String(), authdomain.ActionUpdateAll.String()},
		{admin, platform, authdomain.ResourceEvent.String(), authdomain.ActionDeleteAll.String()},
		{admin, platform, authdomain.ResourceEvent.String(), authdomain.ActionManage.String()},

		{admin, platform, authdomain.ResourcePlatform.String(), authdomain.ActionManage.String()},
		{admin, platform, authdomain.ResourceAnalytics.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourcePayment.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},

		{admin, platform, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourceMember.String(), authdomain.ActionDelete.String()},

		{admin, platform, authdomain.ResourceMedia.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourceMedia.String(), authdomain.ActionDelete.String()},

		{admin, platform, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourceTeam.String(), authdomain.ActionUpdate.String()},
		{admin, platform, authdomain.ResourceTeam.String(), authdomain.ActionDelete.String()},
	}
	policies = append(policies, adminPolicies...)

	// ---- Guest: anonymous public viewing only ----
	guestPolicies := [][]string{
		{guest, platform, authdomain.ResourceEvent.String(), authdomain.ActionRead.String()},
	}
	policies = append(policies, guestPolicies...)

	return policies
}

// GetPlatformRoleHierarchy returns platform-domain role inheritance rules.
func GetPlatformRoleHierarchy() [][]string {
	return [][]string{
		{authdomain.RoleSuperAdmin.String(), authdomain.RoleAdmin.String(), authdomain.DomainPlatform},
	}
}

// ============================================================
// ACCOUNT POLICIES
// ============================================================

// AccountDomainWildcard is the pattern used in policies that apply to
// every account.
const AccountDomainWildcard = "account:*"

// GetAccountPolicies returns the account-domain policy set, composed
// from one block per role.
func GetAccountPolicies() [][]string {
	var policies [][]string
	policies = append(policies, accountAdminPolicies()...)
	policies = append(policies, trainerPolicies()...)
	policies = append(policies, learnerPolicies()...)
	// Guest has NO account-domain policies. Anonymous public browsing
	// is handled entirely by the platform guest grant.
	return policies
}

// ============================================================
// ACCOUNT ADMIN
// ============================================================
//
// Full account and team management. No scoping — admins operate on
// every row in the account.
func accountAdminPolicies() [][]string {
	role := authdomain.RoleAccountAdmin.String()
	d := AccountDomainWildcard

	return [][]string{
		// ---- Account ----
		{role, d, authdomain.ResourceAccount.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceAccount.String(), authdomain.ActionUpdate.String()},
		{role, d, authdomain.ResourceAccount.String(), authdomain.ActionDelete.String()},

		// ---- Members ----
		{role, d, authdomain.ResourceMember.String(), authdomain.ActionCreate.String()},
		{role, d, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceMember.String(), authdomain.ActionUpdate.String()},
		{role, d, authdomain.ResourceMember.String(), authdomain.ActionDelete.String()},
		{role, d, authdomain.ResourceMember.String(), authdomain.ActionInvite.String()},
		{role, d, authdomain.ResourceMember.String(), authdomain.ActionLeave.String()},

		// ---- Teams ----
		{role, d, authdomain.ResourceTeam.String(), authdomain.ActionCreate.String()},
		{role, d, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceTeam.String(), authdomain.ActionUpdate.String()},
		{role, d, authdomain.ResourceTeam.String(), authdomain.ActionDelete.String()},
		{role, d, authdomain.ResourceTeam.String(), authdomain.ActionManage.String()},


				// ---- Registrations ----
		{role, d, authdomain.ResourceRegistration.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceRegistration.String(), authdomain.ActionUpdate.String()},
		{role, d, authdomain.ResourceRegistration.String(), authdomain.ActionDelete.String()},
		{role, d, authdomain.ResourceRegistration.String(), authdomain.ActionExport.String()},

		// ---- Events ----
		{role, d, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String()},
		{role, d, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{role, d, authdomain.ResourceEvent.String(), authdomain.ActionUpdateAll.String()},
		{role, d, authdomain.ResourceEvent.String(), authdomain.ActionDeleteAll.String()},
		{role, d, authdomain.ResourceEvent.String(), authdomain.ActionPublishAll.String()},
		{role, d, authdomain.ResourceEvent.String(), authdomain.ActionManage.String()},

		// ---- Certificates ----
		{role, d, authdomain.ResourceCertificate.String(), authdomain.ActionCreate.String()},
		{role, d, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceCertificate.String(), authdomain.ActionUpdate.String()},
		{role, d, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String()},
		{role, d, authdomain.ResourceCertificate.String(), authdomain.ActionDelete.String()},

		// ---- Attendees ----
		{role, d, authdomain.ResourceAttendee.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceAttendee.String(), authdomain.ActionUpdate.String()},
		{role, d, authdomain.ResourceAttendee.String(), authdomain.ActionExport.String()},

		// ---- Payments ----
		{role, d, authdomain.ResourcePayment.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourcePayment.String(), authdomain.ActionCreate.String()},
		{role, d, authdomain.ResourcePayment.String(), authdomain.ActionRefund.String()},

		// ---- Billing ----
		{role, d, authdomain.ResourceBilling.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceBilling.String(), authdomain.ActionUpdate.String()},

		// ---- Settings ----
		{role, d, authdomain.ResourceSetting.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceSetting.String(), authdomain.ActionUpdate.String()},

		// ---- Dashboard, Profile, Analytics ----
		{role, d, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceProfile.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceProfile.String(), authdomain.ActionUpdate.String()},
		{role, d, authdomain.ResourceAnalytics.String(), authdomain.ActionRead.String()},
	}
}

// ============================================================
// TRAINER
// ============================================================
//
// Operational staff. Sees the whole event roster of the account
// (read_all), but manages only their own events.
//
// Payments are NOT granted. A trainer is an invited worker in someone
// else's account; the money side of that account is the owner's
// concern. If a user owns an account they hold account_admin there,
// which is where they see payments.
//
// Member visibility is account-wide (member:read). Team-scoped
// filtering is applied by the team service for
// GET /teams/:id/members.
func trainerPolicies() [][]string {
	role := authdomain.RoleTrainer.String()
	d := AccountDomainWildcard

	return [][]string{
		// ---- Account (read only) ----
		{role, d, authdomain.ResourceAccount.String(), authdomain.ActionRead.String()},

		// ---- Members (view roster, invite, leave) ----
		{role, d, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceMember.String(), authdomain.ActionInvite.String()},
		{role, d, authdomain.ResourceMember.String(), authdomain.ActionLeave.String()},

		// ---- Teams (read only) ----
		{role, d, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},

		// ---- Events: account-wide read, own-only write ----
		{role, d, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String()},
		{role, d, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{role, d, authdomain.ResourceEvent.String(), authdomain.ActionUpdateOwn.String()},
		{role, d, authdomain.ResourceEvent.String(), authdomain.ActionDeleteOwn.String()},
		{role, d, authdomain.ResourceEvent.String(), authdomain.ActionPublishOwn.String()},

		// ---- Attendees: own-events only ----
		{role, d, authdomain.ResourceAttendee.String(), authdomain.ActionCreate.String()},
		{role, d, authdomain.ResourceAttendee.String(), authdomain.ActionReadOwn.String()},
		{role, d, authdomain.ResourceAttendee.String(), authdomain.ActionUpdateOwn.String()},
		{role, d, authdomain.ResourceAttendee.String(), authdomain.ActionExportOwn.String()},


				// ---- Registrations: own events only ----
		{role, d, authdomain.ResourceRegistration.String(), authdomain.ActionReadOwn.String()},
		{role, d, authdomain.ResourceRegistration.String(), authdomain.ActionUpdateOwn.String()},
		{role, d, authdomain.ResourceRegistration.String(), authdomain.ActionExportOwn.String()},

		// ---- Certificates ----
		{role, d, authdomain.ResourceCertificate.String(), authdomain.ActionCreate.String()},
		{role, d, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String()},


		// ---- Dashboard, Profile, Analytics ----
		{role, d, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceProfile.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceProfile.String(), authdomain.ActionUpdate.String()},
		{role, d, authdomain.ResourceAnalytics.String(), authdomain.ActionRead.String()},
	}
}

// ============================================================
// LEARNER
// ============================================================
//
// Consumes training content. No management anywhere in the account.
// All visibility is scoped by the service layer:
//
//   - Events: only events on teams the caller belongs to.
//   - Attendees: only the caller's own records.
//   - Certificates: only the caller's own certificates.
//   - Team rosters: read via the team members endpoint
//     (GET /teams/:id/members), which the team service scopes to
//     teams the caller is an active member of.
//
// The learner does NOT hold `member:read` or `account:read`. The
// account-level member list is not visible to learners.
//
// The UserMenu role badge reads from the self-service
// GET /accounts/:id/members/me endpoint, which requires no
// member-list grant and returns only the caller's own membership row.
func learnerPolicies() [][]string {
	role := authdomain.RoleLearner.String()
	d := AccountDomainWildcard

	return [][]string{
		// ---- Teams (read only) ----
		{role, d, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},

		// ---- Members ----
		// read_own: read the member roster of a team the caller is a
		// member of. The team service enforces the team-membership
		// check before returning the roster.
		// leave: remove themselves from an account.
		{role, d, authdomain.ResourceMember.String(), authdomain.ActionReadOwn.String()},
		{role, d, authdomain.ResourceMember.String(), authdomain.ActionLeave.String()},

		// ---- Events: visible-scope read + register ----
		{role, d, authdomain.ResourceEvent.String(), authdomain.ActionReadOwn.String()},
		{role, d, authdomain.ResourceEvent.String(), authdomain.ActionRegister.String()},


				// ---- Registrations: own records only ----
		{role, d, authdomain.ResourceRegistration.String(), authdomain.ActionReadOwn.String()},

		// ---- Attendee: own records only ----
		{role, d, authdomain.ResourceAttendee.String(), authdomain.ActionReadOwn.String()},
		{role, d, authdomain.ResourceAttendee.String(), authdomain.ActionUpdateOwn.String()},

		// ---- Certificates: own records only ----
		{role, d, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceCertificate.String(), authdomain.ActionDownload.String()},

		// ---- Profile ----
		{role, d, authdomain.ResourceProfile.String(), authdomain.ActionRead.String()},
		{role, d, authdomain.ResourceProfile.String(), authdomain.ActionUpdate.String()},

		// ---- Dashboard ----
		{role, d, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},
	}
}

// GetAccountRoleHierarchy returns the role hierarchy for account domains.
func GetAccountRoleHierarchy() [][]string {
	return [][]string{
		{authdomain.RoleAccountAdmin.String(), authdomain.RoleTrainer.String(), AccountDomainWildcard},
	}
}