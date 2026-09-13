// internal/modules/auth/authorization/policies.go

package authorization

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"

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

// ============================================================
// PLATFORM POLICIES
// ============================================================

// GetPlatformPolicies returns policies for the "platform" domain.
// Used only by Nuruvent staff: super_admin, admin, and the anonymous
// guest role for public event viewing.
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
		// Users
		{admin, platform, authdomain.ResourceUser.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourceUser.String(), authdomain.ActionUpdate.String()},
		{admin, platform, authdomain.ResourceUser.String(), authdomain.ActionDelete.String()},

		// Institutions
		{admin, platform, authdomain.ResourceInstitution.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourceInstitution.String(), authdomain.ActionUpdate.String()},
		{admin, platform, authdomain.ResourceInstitution.String(), authdomain.ActionDelete.String()},

		// Accounts
		{admin, platform, authdomain.ResourceAccount.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourceAccount.String(), authdomain.ActionUpdate.String()},
		{admin, platform, authdomain.ResourceAccount.String(), authdomain.ActionDelete.String()},

		// Events (cross-tenant)
		{admin, platform, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{admin, platform, authdomain.ResourceEvent.String(), authdomain.ActionUpdateAll.String()},
		{admin, platform, authdomain.ResourceEvent.String(), authdomain.ActionDeleteAll.String()},
		{admin, platform, authdomain.ResourceEvent.String(), authdomain.ActionManage.String()},

		// Platform
		{admin, platform, authdomain.ResourcePlatform.String(), authdomain.ActionManage.String()},

		// Analytics
		{admin, platform, authdomain.ResourceAnalytics.String(), authdomain.ActionRead.String()},

		// Payments (read-only)
		{admin, platform, authdomain.ResourcePayment.String(), authdomain.ActionRead.String()},

		// Certificates (read-only)
		{admin, platform, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},

		// Members (read + remove)
		{admin, platform, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourceMember.String(), authdomain.ActionDelete.String()},

		// Media
		{admin, platform, authdomain.ResourceMedia.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourceMedia.String(), authdomain.ActionDelete.String()},

		// Teams
		{admin, platform, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},
		{admin, platform, authdomain.ResourceTeam.String(), authdomain.ActionUpdate.String()},
		{admin, platform, authdomain.ResourceTeam.String(), authdomain.ActionDelete.String()},
	}
	policies = append(policies, adminPolicies...)

	// ---- Guest: anonymous public viewing only ----
	// Platform-domain guest. Distinct from the account-domain guest role
	// (see GetAccountPolicies). Platform guest = unauthenticated visitor.
	// Account guest = invited external collaborator.
	guestPolicies := [][]string{
		{guest, platform, authdomain.ResourceEvent.String(), authdomain.ActionRead.String()},
	}
	policies = append(policies, guestPolicies...)

	return policies
}

// GetPlatformRoleHierarchy returns platform-domain role inheritance rules.
//
// super_admin inherits admin's capabilities within the platform domain.
// This is a role-to-role rule, not a user-to-role assignment.
func GetPlatformRoleHierarchy() [][]string {
	return [][]string{
		{authdomain.RoleSuperAdmin.String(), authdomain.RoleAdmin.String(), authdomain.DomainPlatform},
	}
}

// ============================================================
// ACCOUNT POLICIES
// ============================================================

// AccountDomainWildcard is the pattern used in policies that apply to
// every account. At check time, keyMatch2 resolves it to any concrete
// "account:<uuid>" domain.
const AccountDomainWildcard = "account:*"

// GetAccountPolicies returns the account-domain policy set.
//
// Domain: "account:*" (wildcard, matches every account)
//
// Applies to both personal and institution accounts. The account type
// distinction is product metadata (billing, onboarding, KYC); it does
// not affect authorization.
//
// Teams are NOT authz domains. The policies below cover both account-level
// and team-level actions, since team membership is enforced as data
// (team_members) in the service layer.
func GetAccountPolicies() [][]string {
	accountAdmin := authdomain.RoleAccountAdmin.String()
	trainer := authdomain.RoleTrainer.String()
	guest := authdomain.RoleGuest.String()
	domain := AccountDomainWildcard

	var policies [][]string

	// ============================================================
	// ACCOUNT ADMIN - Full account and team management
	// ============================================================
	accountAdminPolicies := [][]string{
		// ---- Account ----
		{accountAdmin, domain, authdomain.ResourceAccount.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceAccount.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceAccount.String(), authdomain.ActionDelete.String()},

		// ---- Members ----
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionDelete.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionInvite.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionLeave.String()},

		// ---- Teams ----
		{accountAdmin, domain, authdomain.ResourceTeam.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceTeam.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceTeam.String(), authdomain.ActionDelete.String()},
		{accountAdmin, domain, authdomain.ResourceTeam.String(), authdomain.ActionManage.String()},

		// ---- Events ----
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdateAll.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionDeleteAll.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionPublishAll.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionManage.String()},

		// ---- Certificates ----
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String()},
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionDelete.String()},

		// ---- Attendees ----
		{accountAdmin, domain, authdomain.ResourceAttendee.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceAttendee.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceAttendee.String(), authdomain.ActionExport.String()},

		// ---- Payments ----
		{accountAdmin, domain, authdomain.ResourcePayment.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourcePayment.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourcePayment.String(), authdomain.ActionRefund.String()},

		// ---- Billing ----
		{accountAdmin, domain, authdomain.ResourceBilling.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceBilling.String(), authdomain.ActionUpdate.String()},

		// ---- Settings ----
		{accountAdmin, domain, authdomain.ResourceSetting.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceSetting.String(), authdomain.ActionUpdate.String()},

		// ---- Dashboard, Profile, Analytics ----
		{accountAdmin, domain, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceProfile.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceProfile.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceAnalytics.String(), authdomain.ActionRead.String()},
	}
	policies = append(policies, accountAdminPolicies...)

	// ============================================================
	// TRAINER - Training-focused, no account/team management
	// ============================================================
	trainerPolicies := [][]string{
		// ---- Account (read only) ----
		{trainer, domain, authdomain.ResourceAccount.String(), authdomain.ActionRead.String()},

		// ---- Members (view roster, leave) ----
		{trainer, domain, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
		{trainer, domain, authdomain.ResourceMember.String(), authdomain.ActionLeave.String()},

		// ---- Teams (read only) ----
		{trainer, domain, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},

		// ---- Events ----
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdateAll.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionPublishAll.String()},

		// ---- Attendees ----
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionCreate.String()},
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionRead.String()},
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionUpdate.String()},
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionDelete.String()},
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionExport.String()},

		// ---- Certificates ----
		{trainer, domain, authdomain.ResourceCertificate.String(), authdomain.ActionCreate.String()},
		{trainer, domain, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{trainer, domain, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String()},

		// ---- Payments (read only) ----
		{trainer, domain, authdomain.ResourcePayment.String(), authdomain.ActionRead.String()},

		// ---- Dashboard, Profile, Analytics ----
		{trainer, domain, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},
		{trainer, domain, authdomain.ResourceProfile.String(), authdomain.ActionRead.String()},
		{trainer, domain, authdomain.ResourceProfile.String(), authdomain.ActionUpdate.String()},
		{trainer, domain, authdomain.ResourceAnalytics.String(), authdomain.ActionRead.String()},
	}
	policies = append(policies, trainerPolicies...)

	// ============================================================
	// GUEST - Narrow cross-account collaboration
	// ============================================================
	// Account-domain guest: invited external collaborator with minimal
	// permissions. Distinct from the platform-domain guest, which is for
	// anonymous public viewing.
	guestPolicies := [][]string{
		{guest, domain, authdomain.ResourceAccount.String(), authdomain.ActionRead.String()},
		{guest, domain, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},
		{guest, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{guest, domain, authdomain.ResourceProfile.String(), authdomain.ActionRead.String()},
		{guest, domain, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},
	}
	policies = append(policies, guestPolicies...)

	return policies
}

// GetAccountRoleHierarchy returns the role hierarchy for account domains.
//
// Domain: "account:*" (wildcard)
//
// account_admin inherits everything trainer has, within the same account
// domain. This is a role-to-role rule, not a user assignment. It applies
// to every account but only grants inheritance within a single account
// at check time.
//
// Note: user role assignments (from account_members) must ALWAYS use
// concrete domains ("account:<uuid>"), never the wildcard.
func GetAccountRoleHierarchy() [][]string {
	return [][]string{
		{authdomain.RoleAccountAdmin.String(), authdomain.RoleTrainer.String(), AccountDomainWildcard},
	}
}