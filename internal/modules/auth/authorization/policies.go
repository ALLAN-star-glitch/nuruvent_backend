// internal/modules/authorization/policies.go

package authorization

// GetAccountAdminPolicies returns policies for account_admin (full account management)
func GetAccountAdminPolicies(domain string) [][]string {
	resources := []Resource{
		ResourceEvent,
		ResourceCertificate,
		ResourceAttendee,
		ResourcePayment,
		ResourcePayout,
		ResourceInstitution,
		ResourceMember,
		ResourceAccount,
		ResourceProfile,
		ResourceDashboard,
		ResourceAnalytics,
		ResourceNotification,
		ResourceMedia,
	}

	actions := []Action{
		ActionCreate,
		ActionRead,
		ActionUpdate,
		ActionDelete,
	}

	var policies [][]string
	for _, resource := range resources {
		for _, action := range actions {
			policies = append(policies, []string{
				RoleAccountAdmin.String(),
				domain,
				resource.String(),
				action.String(),
			})
		}
	}

	// Add special actions
	specialPolicies := [][]string{
		{RoleAccountAdmin.String(), domain, ResourceCertificate.String(), ActionIssue.String()},
		{RoleAccountAdmin.String(), domain, ResourcePayment.String(), ActionRefund.String()},
		{RoleAccountAdmin.String(), domain, ResourceAttendee.String(), ActionExport.String()},
		{RoleAccountAdmin.String(), domain, ResourceEvent.String(), ActionManage.String()},
		{RoleAccountAdmin.String(), domain, ResourceInstitution.String(), ActionManage.String()},
		{RoleAccountAdmin.String(), domain, ResourceMember.String(), ActionManage.String()},
		{RoleAccountAdmin.String(), domain, ResourceMember.String(), ActionInvite.String()},
		{RoleAccountAdmin.String(), domain, ResourcePayout.String(), ActionManage.String()},
		{RoleAccountAdmin.String(), domain, ResourceAccount.String(), ActionManage.String()},
		{RoleAccountAdmin.String(), domain, ResourceProfile.String(), ActionManage.String()},
	}

	policies = append(policies, specialPolicies...)
	return policies
}

// GetEventManagerPolicies returns policies for event_manager
func GetEventManagerPolicies(domain string) [][]string {
	return [][]string{
		// Event management
		{RoleEventManager.String(), domain, ResourceEvent.String(), ActionCreate.String()},
		{RoleEventManager.String(), domain, ResourceEvent.String(), ActionRead.String()},
		{RoleEventManager.String(), domain, ResourceEvent.String(), ActionUpdate.String()},
		{RoleEventManager.String(), domain, ResourceEvent.String(), ActionDelete.String()},
		{RoleEventManager.String(), domain, ResourceEvent.String(), ActionManage.String()},

		// Certificate management
		{RoleEventManager.String(), domain, ResourceCertificate.String(), ActionCreate.String()},
		{RoleEventManager.String(), domain, ResourceCertificate.String(), ActionRead.String()},
		{RoleEventManager.String(), domain, ResourceCertificate.String(), ActionIssue.String()},

		// Attendee management
		{RoleEventManager.String(), domain, ResourceAttendee.String(), ActionRead.String()},
		{RoleEventManager.String(), domain, ResourceAttendee.String(), ActionUpdate.String()},
		{RoleEventManager.String(), domain, ResourceAttendee.String(), ActionExport.String()},

		// Payment - Read only
		{RoleEventManager.String(), domain, ResourcePayment.String(), ActionRead.String()},

		// Dashboard
		{RoleEventManager.String(), domain, ResourceDashboard.String(), ActionRead.String()},
		{RoleEventManager.String(), domain, ResourceAnalytics.String(), ActionRead.String()},

		// Media
		{RoleEventManager.String(), domain, ResourceMedia.String(), ActionCreate.String()},
		{RoleEventManager.String(), domain, ResourceMedia.String(), ActionRead.String()},
		{RoleEventManager.String(), domain, ResourceMedia.String(), ActionUpdate.String()},
		{RoleEventManager.String(), domain, ResourceMedia.String(), ActionDelete.String()},
	}
}

// GetTeamMemberPolicies returns policies for team_member (view-only)
func GetTeamMemberPolicies(domain string) [][]string {
	return [][]string{
		{RoleTeamMember.String(), domain, ResourceEvent.String(), ActionRead.String()},
		{RoleTeamMember.String(), domain, ResourceCertificate.String(), ActionRead.String()},
		{RoleTeamMember.String(), domain, ResourceDashboard.String(), ActionRead.String()},
		{RoleTeamMember.String(), domain, ResourceAnalytics.String(), ActionRead.String()},
		{RoleTeamMember.String(), domain, ResourceAttendee.String(), ActionRead.String()},
		{RoleTeamMember.String(), domain, ResourceMedia.String(), ActionRead.String()},
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
		{RoleAccountAdmin.String(), RoleEventManager.String(), domain},
		{RoleAccountAdmin.String(), RoleTeamMember.String(), domain},

		// Event Manager inherits team member
		{RoleEventManager.String(), RoleTeamMember.String(), domain},
	}
}

// GetPlatformPolicies returns platform-level policies
func GetPlatformPolicies() [][]string {
	var allPolicies [][]string

	// Admin policies
	adminPolicies := [][]string{
		{RoleAdmin.String(), DomainPlatform, ResourceAccount.String(), ActionRead.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceAccount.String(), ActionUpdate.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceInstitution.String(), ActionRead.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceEvent.String(), ActionRead.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceEvent.String(), ActionUpdate.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceEvent.String(), ActionDelete.String()},
		{RoleAdmin.String(), DomainPlatform, ResourcePlatform.String(), ActionManage.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceAnalytics.String(), ActionRead.String()},
		{RoleAdmin.String(), DomainPlatform, ResourcePayment.String(), ActionRead.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceCertificate.String(), ActionRead.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceMember.String(), ActionRead.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceMember.String(), ActionDelete.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceMedia.String(), ActionRead.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceMedia.String(), ActionDelete.String()},
	}
	allPolicies = append(allPolicies, adminPolicies...)

	// Super admin policies (full access to everything)
	superAdminPolicies := [][]string{
		{RoleSuperAdmin.String(), DomainPlatform, ResourcePlatform.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceAccount.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceInstitution.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceEvent.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceCertificate.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceAttendee.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourcePayment.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourcePayout.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceMember.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceProfile.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceDashboard.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceAnalytics.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceNotification.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceMedia.String(), ActionManage.String()},
	}
	allPolicies = append(allPolicies, superAdminPolicies...)

	// Guest policies (public access)
	guestPolicies := [][]string{
		{RoleGuest.String(), DomainPlatform, ResourceEvent.String(), ActionRead.String()},
	}
	allPolicies = append(allPolicies, guestPolicies...)

	return allPolicies
}

// GetPlatformRoleHierarchy returns role inheritance rules for platform domain
func GetPlatformRoleHierarchy() [][]string {
	return [][]string{
		// Super Admin inherits Admin
		{RoleSuperAdmin.String(), RoleAdmin.String(), DomainPlatform},

		// Admin inherits Account Admin (platform admin can act as account admin)
		{RoleAdmin.String(), RoleAccountAdmin.String(), DomainPlatform},
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