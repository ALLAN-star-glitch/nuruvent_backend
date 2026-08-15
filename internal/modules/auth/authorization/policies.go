// internal/modules/auth/authorization/policies.go

package authorization

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"

// GetAccountAdminPolicies returns policies for account_admin (full account management)
func GetAccountAdminPolicies(domain string) [][]string {
	resources := []authdomain.Resource{
		authdomain.ResourceEvent,
		authdomain.ResourceCertificate,
		authdomain.ResourceAttendee,
		authdomain.ResourcePayment,
		authdomain.ResourcePayout,
		authdomain.ResourceInstitution,
		authdomain.ResourceMember,
		authdomain.ResourceAccount,
		authdomain.ResourceProfile,
		authdomain.ResourceDashboard,
		authdomain.ResourceAnalytics,
		authdomain.ResourceNotification,
		authdomain.ResourceMedia,
	}

	actions := []authdomain.Action{
		authdomain.ActionCreate,
		authdomain.ActionRead,
		authdomain.ActionUpdate,
		authdomain.ActionDelete,
	}

	var policies [][]string
	for _, resource := range resources {
		for _, action := range actions {
			policies = append(policies, []string{
				authdomain.RoleAccountAdmin.String(),
				domain,
				resource.String(),
				action.String(),
			})
		}
	}

	// Add special actions
	specialPolicies := [][]string{
		{authdomain.RoleAccountAdmin.String(), domain, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String()},
		{authdomain.RoleAccountAdmin.String(), domain, authdomain.ResourcePayment.String(), authdomain.ActionRefund.String()},
		{authdomain.RoleAccountAdmin.String(), domain, authdomain.ResourceAttendee.String(), authdomain.ActionExport.String()},
		{authdomain.RoleAccountAdmin.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionManage.String()},
		{authdomain.RoleAccountAdmin.String(), domain, authdomain.ResourceInstitution.String(), authdomain.ActionManage.String()},
		{authdomain.RoleAccountAdmin.String(), domain, authdomain.ResourceMember.String(), authdomain.ActionManage.String()},
		{authdomain.RoleAccountAdmin.String(), domain, authdomain.ResourceMember.String(), authdomain.ActionInvite.String()},
		{authdomain.RoleAccountAdmin.String(), domain, authdomain.ResourcePayout.String(), authdomain.ActionManage.String()},
		{authdomain.RoleAccountAdmin.String(), domain, authdomain.ResourceAccount.String(), authdomain.ActionManage.String()},
		{authdomain.RoleAccountAdmin.String(), domain, authdomain.ResourceProfile.String(), authdomain.ActionManage.String()},
	}

	policies = append(policies, specialPolicies...)
	return policies
}

// GetEventManagerPolicies returns policies for event_manager
func GetEventManagerPolicies(domain string) [][]string {
	return [][]string{
		// Event management
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionRead.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdate.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionDelete.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionManage.String()},

		// Certificate management
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceCertificate.String(), authdomain.ActionCreate.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String()},

		// Attendee management
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceAttendee.String(), authdomain.ActionRead.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceAttendee.String(), authdomain.ActionUpdate.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceAttendee.String(), authdomain.ActionExport.String()},

		// Payment - Read only
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourcePayment.String(), authdomain.ActionRead.String()},

		// Dashboard
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceAnalytics.String(), authdomain.ActionRead.String()},

		// Media
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceMedia.String(), authdomain.ActionCreate.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceMedia.String(), authdomain.ActionRead.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceMedia.String(), authdomain.ActionUpdate.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceMedia.String(), authdomain.ActionDelete.String()},
	}
}

// GetTeamMemberPolicies returns policies for team_member (view-only)
func GetTeamMemberPolicies(domain string) [][]string {
	return [][]string{
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionRead.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceAnalytics.String(), authdomain.ActionRead.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceAttendee.String(), authdomain.ActionRead.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceMedia.String(), authdomain.ActionRead.String()},
	}
}

// GetPersonalAccountPolicies returns policies for personal account type
// Personal accounts can do everything an individual can (attend AND host)
func GetPersonalAccountPolicies(domain string) [][]string {
	// Personal accounts get account_admin capabilities for their own account
	return GetAccountAdminPolicies(domain)
}

// GetAccountPolicies returns all account policies for a domain
func GetAccountPolicies(domain string) [][]string {
	var allPolicies [][]string

	allPolicies = append(allPolicies, GetAccountAdminPolicies(domain)...)
	allPolicies = append(allPolicies, GetEventManagerPolicies(domain)...)
	allPolicies = append(allPolicies, GetTeamMemberPolicies(domain)...)

	return allPolicies
}

// GetAccountRoleHierarchy returns role inheritance rules for an account domain
func GetAccountRoleHierarchy(domain string) [][]string {
	return [][]string{
		// Account Admin inherits all account roles
		{authdomain.RoleAccountAdmin.String(), authdomain.RoleEventManager.String(), domain},
		{authdomain.RoleAccountAdmin.String(), authdomain.RoleTeamMember.String(), domain},

		// Event Manager inherits team member
		{authdomain.RoleEventManager.String(), authdomain.RoleTeamMember.String(), domain},
	}
}

// GetPlatformPolicies returns platform-level policies
func GetPlatformPolicies() [][]string {
	var allPolicies [][]string

	// Admin policies
	adminPolicies := [][]string{
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceAccount.String(), authdomain.ActionRead.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceAccount.String(), authdomain.ActionUpdate.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceInstitution.String(), authdomain.ActionRead.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceEvent.String(), authdomain.ActionRead.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceEvent.String(), authdomain.ActionUpdate.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceEvent.String(), authdomain.ActionDelete.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourcePlatform.String(), authdomain.ActionManage.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceAnalytics.String(), authdomain.ActionRead.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourcePayment.String(), authdomain.ActionRead.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceMember.String(), authdomain.ActionRead.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceMember.String(), authdomain.ActionDelete.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceMedia.String(), authdomain.ActionRead.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceMedia.String(), authdomain.ActionDelete.String()},
	}
	allPolicies = append(allPolicies, adminPolicies...)

	// Super admin policies (full access to everything)
	superAdminPolicies := [][]string{
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourcePlatform.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceAccount.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceInstitution.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceEvent.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceCertificate.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceAttendee.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourcePayment.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourcePayout.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceMember.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceProfile.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceDashboard.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceAnalytics.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceNotification.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceMedia.String(), authdomain.ActionManage.String()},
	}
	allPolicies = append(allPolicies, superAdminPolicies...)

	// Guest policies (public access)
	guestPolicies := [][]string{
		{authdomain.RoleGuest.String(), authdomain.DomainPlatform, authdomain.ResourceEvent.String(), authdomain.ActionRead.String()},
	}
	allPolicies = append(allPolicies, guestPolicies...)

	return allPolicies
}

// GetPlatformRoleHierarchy returns role inheritance rules for platform domain
func GetPlatformRoleHierarchy() [][]string {
	return [][]string{
		// Super Admin inherits Admin
		{authdomain.RoleSuperAdmin.String(), authdomain.RoleAdmin.String(), authdomain.DomainPlatform},

		// Admin inherits Account Admin (platform admin can act as account admin)
		{authdomain.RoleAdmin.String(), authdomain.RoleAccountAdmin.String(), authdomain.DomainPlatform},
	}
}

// GetAllPolicies returns all policies for a domain (platform + account)
func GetAllPolicies(domain string) [][]string {
	var allPolicies [][]string

	allPolicies = append(allPolicies, GetPlatformPolicies()...)
	allPolicies = append(allPolicies, GetAccountPolicies(domain)...)

	return allPolicies
}

// GetAllRoleHierarchy returns all role inheritance rules
func GetAllRoleHierarchy(domain string) [][]string {
	var allHierarchy [][]string

	allHierarchy = append(allHierarchy, GetPlatformRoleHierarchy()...)
	allHierarchy = append(allHierarchy, GetAccountRoleHierarchy(domain)...)

	return allHierarchy
}