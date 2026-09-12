// internal/modules/auth/authorization/policies.go

package authorization

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"

// ============================================================
// PERSONAL TEAM POLICIES
// ============================================================

// GetPersonalTeamPolicies returns all policies for a personal team
// domain should be "personal:team:{user_id}"
func GetPersonalTeamPolicies(domain string) [][]string {
	accountAdmin := authdomain.RoleAccountAdmin.String()
	trainer := authdomain.RoleTrainer.String()

	var policies [][]string

	// ============================================================
	// ACCOUNT ADMIN - Full access (ALL)
	// ============================================================
	accountAdminPolicies := [][]string{
		// Event permissions - ALL
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdateAll.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionDeleteAll.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionPublishAll.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionManage.String()},

		// Certificate permissions
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String()},
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionDelete.String()},

		// Attendee permissions
		{accountAdmin, domain, authdomain.ResourceAttendee.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceAttendee.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceAttendee.String(), authdomain.ActionExport.String()},

		// Payment permissions
		{accountAdmin, domain, authdomain.ResourcePayment.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourcePayment.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourcePayment.String(), authdomain.ActionRefund.String()},

		// Member permissions
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionDelete.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionInvite.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionLeave.String()},   // ← added

		// Institution permissions (limited for personal)
		{accountAdmin, domain, authdomain.ResourceInstitution.String(), authdomain.ActionRead.String()},

		// Team permissions
		{accountAdmin, domain, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceTeam.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceTeam.String(), authdomain.ActionDelete.String()},

		// Dashboard and profile
		{accountAdmin, domain, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceProfile.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceProfile.String(), authdomain.ActionReadAll.String()},
		{accountAdmin, domain, authdomain.ResourceProfile.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceProfile.String(), authdomain.ActionUpdateAll.String()},
	}
	policies = append(policies, accountAdminPolicies...)

	// ============================================================
	// TRAINER - Training focused (ALL)
	// ============================================================
	trainerPolicies := [][]string{
		// Event permissions - ALL
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdateAll.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionDeleteAll.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionPublishAll.String()},

		// Attendee permissions
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionCreate.String()},
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionRead.String()},
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionUpdate.String()},
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionDelete.String()},
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionExport.String()},

		// Certificate permissions
		{trainer, domain, authdomain.ResourceCertificate.String(), authdomain.ActionCreate.String()},
		{trainer, domain, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{trainer, domain, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String()},
		{trainer, domain, authdomain.ResourceCertificate.String(), authdomain.ActionDelete.String()},

		// Payment - Read only
		{trainer, domain, authdomain.ResourcePayment.String(), authdomain.ActionRead.String()},

		// Dashboard
		{trainer, domain, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},

		// Team - Read only
		{trainer, domain, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},
	}
	policies = append(policies, trainerPolicies...)

	return policies
}

// ============================================================
// INSTITUTION TEAM POLICIES
// ============================================================

// GetInstitutionTeamPolicies returns all policies for an institution team
// domain should be "institution:team:{institution_id}"
func GetInstitutionTeamPolicies(domain string) [][]string {
	accountAdmin := authdomain.RoleAccountAdmin.String()
	trainer := authdomain.RoleTrainer.String()

	var policies [][]string

	// ============================================================
	// ACCOUNT ADMIN - Full access (ALL)
	// ============================================================
	accountAdminPolicies := [][]string{
		// Event permissions - ALL
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdateAll.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionDeleteAll.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionPublishAll.String()},
		{accountAdmin, domain, authdomain.ResourceEvent.String(), authdomain.ActionManage.String()},

		// Certificate permissions
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String()},
		{accountAdmin, domain, authdomain.ResourceCertificate.String(), authdomain.ActionDelete.String()},

		// Attendee permissions
		{accountAdmin, domain, authdomain.ResourceAttendee.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceAttendee.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceAttendee.String(), authdomain.ActionExport.String()},

		// Payment permissions
		{accountAdmin, domain, authdomain.ResourcePayment.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourcePayment.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourcePayment.String(), authdomain.ActionRefund.String()},

		// Member permissions
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionDelete.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionInvite.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionLeave.String()},   // ← added

		// Institution permissions (full for institution)
		{accountAdmin, domain, authdomain.ResourceInstitution.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceInstitution.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceInstitution.String(), authdomain.ActionManage.String()},

		// Team permissions
		{accountAdmin, domain, authdomain.ResourceTeam.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceTeam.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceTeam.String(), authdomain.ActionDelete.String()},

		// Dashboard and profile
		{accountAdmin, domain, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceProfile.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceProfile.String(), authdomain.ActionReadAll.String()},
		{accountAdmin, domain, authdomain.ResourceProfile.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceProfile.String(), authdomain.ActionUpdateAll.String()},
	}
	policies = append(policies, accountAdminPolicies...)

	// ============================================================
	// TRAINER - Training focused (ALL)
	// ============================================================
	trainerPolicies := [][]string{
		// Event permissions - ALL
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdateAll.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionDeleteAll.String()},
		{trainer, domain, authdomain.ResourceEvent.String(), authdomain.ActionPublishAll.String()},

		// Attendee permissions
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionCreate.String()},
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionRead.String()},
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionUpdate.String()},
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionDelete.String()},
		{trainer, domain, authdomain.ResourceAttendee.String(), authdomain.ActionExport.String()},

		// Certificate permissions
		{trainer, domain, authdomain.ResourceCertificate.String(), authdomain.ActionCreate.String()},
		{trainer, domain, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{trainer, domain, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String()},
		{trainer, domain, authdomain.ResourceCertificate.String(), authdomain.ActionDelete.String()},

		// Payment - Read only
		{trainer, domain, authdomain.ResourcePayment.String(), authdomain.ActionRead.String()},

		// Dashboard
		{trainer, domain, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},

		// Team - Read only
		{trainer, domain, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},

		// Member — view roster and leave
		{trainer, domain, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},   // ← added
		{trainer, domain, authdomain.ResourceMember.String(), authdomain.ActionLeave.String()},  // ← added
	}
	policies = append(policies, trainerPolicies...)

	return policies
}

// ============================================================
// ACCOUNT POLICIES (NEW)
// ============================================================

// GetAccountPolicies returns all policies for an account domain
// domain should be "account:{account_id}"
// ============================================================
// ACCOUNT POLICIES
// ============================================================

// GetAccountPolicies returns all policies for an account domain.
// Domain format: "account:{account_id}".
//
// NOTE: This is the domain used by /accounts/:id/members routes.
// The member:* grants below are what actually authorize those
// endpoints — the team policy functions do NOT cover them.
func GetAccountPolicies(domain string) [][]string {
	accountAdmin := authdomain.RoleAccountAdmin.String()
	trainer := authdomain.RoleTrainer.String()

	var policies [][]string

	// ============================================================
	// ACCOUNT ADMIN — Full account management
	// ============================================================
	accountAdminPolicies := [][]string{
		// ---- Account ----
		{accountAdmin, domain, authdomain.ResourceAccount.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceAccount.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceAccount.String(), authdomain.ActionDelete.String()},
		{accountAdmin, domain, authdomain.ResourceAccount.String(), authdomain.ActionMemberAdd.String()},
		{accountAdmin, domain, authdomain.ResourceAccount.String(), authdomain.ActionMemberRemove.String()},

		// ---- Members ----
		// Members resolve to domain=account:<id>, resource=member.
		// Without these grants, /accounts/:id/members returns 403.
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionCreate.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionUpdate.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionDelete.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionInvite.String()},
		{accountAdmin, domain, authdomain.ResourceMember.String(), authdomain.ActionLeave.String()},

		// ---- Billing ----
		{accountAdmin, domain, authdomain.ResourceBilling.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceBilling.String(), authdomain.ActionUpdate.String()},

		// ---- Settings ----
		{accountAdmin, domain, authdomain.ResourceSetting.String(), authdomain.ActionRead.String()},
		{accountAdmin, domain, authdomain.ResourceSetting.String(), authdomain.ActionUpdate.String()},
	}
	policies = append(policies, accountAdminPolicies...)

	// ============================================================
	// TRAINER — Limited account access
	// ============================================================
	trainerPolicies := [][]string{
		// Account — Read only
		{trainer, domain, authdomain.ResourceAccount.String(), authdomain.ActionRead.String()},

		// Members — Can view the roster and leave, but not manage it.
		//
		// Leaving is a self-service action: every member must be able
		// to remove themselves. It is deliberately NOT ActionDelete,
		// because "leave" (self) and "delete" (someone else) are
		// distinct privileges — an admin can do both, a trainer only
		// the former.
		{trainer, domain, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
		{trainer, domain, authdomain.ResourceMember.String(), authdomain.ActionLeave.String()},
	}
	policies = append(policies, trainerPolicies...)

	return policies
}

// ============================================================
// TEAM ROLE HIERARCHY
// ============================================================

// GetTeamRoleHierarchy returns the role hierarchy for a team domain
// domain should be "personal:team:{id}" or "institution:team:{id}"
func GetTeamRoleHierarchy(domain string) [][]string {
	return [][]string{
		// Account Admin inherits Trainer
		{authdomain.RoleAccountAdmin.String(), authdomain.RoleTrainer.String(), domain},
	}
}

// ============================================================
// ACCOUNT ROLE HIERARCHY (NEW)
// ============================================================

// GetAccountRoleHierarchy returns the role hierarchy for an account domain
// domain should be "account:{account_id}"
func GetAccountRoleHierarchy(domain string) [][]string {
	return [][]string{
		// Account Admin inherits Trainer
		{authdomain.RoleAccountAdmin.String(), authdomain.RoleTrainer.String(), domain},
	}
}

// ============================================================
// PLATFORM POLICIES
// ============================================================

// GetPlatformPolicies returns platform-level policies
func GetPlatformPolicies() [][]string {
	var allPolicies [][]string

	admin := authdomain.RoleAdmin.String()
	superAdmin := authdomain.RoleSuperAdmin.String()
	guest := authdomain.RoleGuest.String()

	// Admin policies
	adminPolicies := [][]string{
		{admin, authdomain.DomainPlatform, authdomain.ResourceUser.String(), authdomain.ActionRead.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceUser.String(), authdomain.ActionUpdate.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceUser.String(), authdomain.ActionDelete.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceInstitution.String(), authdomain.ActionRead.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceInstitution.String(), authdomain.ActionUpdate.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceInstitution.String(), authdomain.ActionDelete.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceEvent.String(), authdomain.ActionRead.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceEvent.String(), authdomain.ActionUpdate.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceEvent.String(), authdomain.ActionDelete.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourcePlatform.String(), authdomain.ActionManage.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceAnalytics.String(), authdomain.ActionRead.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourcePayment.String(), authdomain.ActionRead.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceMember.String(), authdomain.ActionDelete.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceMedia.String(), authdomain.ActionRead.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceMedia.String(), authdomain.ActionDelete.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceTeam.String(), authdomain.ActionUpdate.String()},
		{admin, authdomain.DomainPlatform, authdomain.ResourceTeam.String(), authdomain.ActionDelete.String()},
	}
	allPolicies = append(allPolicies, adminPolicies...)

	// Super admin policies (full access to everything)
	superAdminPolicies := [][]string{
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourcePlatform.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourceUser.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourceInstitution.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourceEvent.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourceCertificate.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourceAttendee.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourcePayment.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourcePayout.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourceMember.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourceProfile.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourceDashboard.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourceAnalytics.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourceNotification.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourceMedia.String(), authdomain.ActionManage.String()},
		{superAdmin, authdomain.DomainPlatform, authdomain.ResourceTeam.String(), authdomain.ActionManage.String()},
	}
	allPolicies = append(allPolicies, superAdminPolicies...)

	// Guest policies (public access)
	guestPolicies := [][]string{
		{guest, authdomain.DomainPlatform, authdomain.ResourceEvent.String(), authdomain.ActionRead.String()},
	}
	allPolicies = append(allPolicies, guestPolicies...)

	return allPolicies
}

// GetPlatformRoleHierarchy returns role inheritance rules for platform domain
func GetPlatformRoleHierarchy() [][]string {
	return [][]string{
		// Super Admin inherits Admin
		{authdomain.RoleSuperAdmin.String(), authdomain.RoleAdmin.String(), authdomain.DomainPlatform},
	}
}