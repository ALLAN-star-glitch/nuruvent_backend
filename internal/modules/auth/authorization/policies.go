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

// GetAccountPolicies returns the account-domain policy set.
func GetAccountPolicies() [][]string {
	accountAdmin := authdomain.RoleAccountAdmin.String()
	trainer := authdomain.RoleTrainer.String()
	learner := authdomain.RoleLearner.String()
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
	// TRAINER - Training-focused, own-only event management
	// ============================================================
	//
	// Trainers may create events and manage the ones they created.
	// They may not edit or delete events created by other trainers
	// or by account admins. Ownership is enforced by the events
	// service using the *_own action codes below.
	//
	// Trainers may invite new members (as trainer or learner) but
	// cannot invite account_admins — that escalation path is gated
	// in the team service by an explicit CanAddMember check.
	trainerPolicies := [][]string{
		// ---- Account (read only) ----
		{trainer, domain, authdomain.ResourceAccount.String(), authdomain.ActionRead.String()},

		// ---- Members (view roster, invite, leave) ----
		{trainer, domain, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
		{trainer, domain, authdomain.ResourceMember.String(), authdomain.ActionInvite.String()},
		{trainer, domain, authdomain.ResourceMember.String(), authdomain.ActionLeave.String()},

		// ---- Teams (read only) ----
		{trainer, domain, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},

		// ---- Events (own only) ----
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdateOwn.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionDeleteOwn.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionPublishOwn.String()},

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
	// LEARNER - Consumes training content, no management
	// ============================================================
	learnerPolicies := [][]string{
		// ---- Account (read only) ----
		{learner, domain, authdomain.ResourceAccount.String(), authdomain.ActionRead.String()},

		// ---- Teams (read only) ----
		{learner, domain, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},

		// ---- Events (view + register only) ----
		{learner, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{learner, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadOwn.String()},
		{learner, domain, authdomain.ResourceEvent.String(), authdomain.ActionRegister.String()},

		// ---- Attendee (own records) ----
		{learner, domain, authdomain.ResourceAttendee.String(), authdomain.ActionRead.String()},
		{learner, domain, authdomain.ResourceAttendee.String(), authdomain.ActionUpdateOwn.String()},

		// ---- Certificates (view + download own) ----
		{learner, domain, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{learner, domain, authdomain.ResourceCertificate.String(), authdomain.ActionDownload.String()},

		// ---- Profile ----
		{learner, domain, authdomain.ResourceProfile.String(), authdomain.ActionRead.String()},
		{learner, domain, authdomain.ResourceProfile.String(), authdomain.ActionUpdate.String()},

		// ---- Dashboard (view own) ----
		{learner, domain, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},

		{learner, domain, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
	}
	policies = append(policies, learnerPolicies...)

	// ============================================================
	// GUEST - Narrow cross-account collaboration
	// ============================================================
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
func GetAccountRoleHierarchy() [][]string {
	return [][]string{
		{authdomain.RoleAccountAdmin.String(), authdomain.RoleTrainer.String(), AccountDomainWildcard},
	}
}