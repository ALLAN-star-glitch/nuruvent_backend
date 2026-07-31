package authorization

// GetBusinessAdminPolicies returns policies for business_admin (full business management)
func GetBusinessAdminPolicies(domain string) [][]string {
	resources := []Resource{
		ResourceEvent,
		ResourceCertificate,
		ResourceAttendee,
		ResourcePayment,
		ResourcePayout,
		ResourceBusiness,
		ResourceMember,
		ResourceDashboard,
		ResourceAnalytics,
		ResourceNotification,
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
				RoleBusinessAdmin.String(),
				domain,
				resource.String(),
				action.String(),
			})
		}
	}

	// Add special actions
	specialPolicies := [][]string{
		{RoleBusinessAdmin.String(), domain, ResourceCertificate.String(), ActionIssue.String()},
		{RoleBusinessAdmin.String(), domain, ResourcePayment.String(), ActionRefund.String()},
		{RoleBusinessAdmin.String(), domain, ResourceAttendee.String(), ActionExport.String()},
		{RoleBusinessAdmin.String(), domain, ResourceEvent.String(), ActionManage.String()},
		{RoleBusinessAdmin.String(), domain, ResourceBusiness.String(), ActionManage.String()},
		{RoleBusinessAdmin.String(), domain, ResourceMember.String(), ActionManage.String()},
		{RoleBusinessAdmin.String(), domain, ResourcePayout.String(), ActionManage.String()},
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
	}
}

// GetMemberPolicies returns policies for member (read-only)
func GetMemberPolicies(domain string) [][]string {
	return [][]string{
		{RoleMember.String(), domain, ResourceEvent.String(), ActionRead.String()},
		{RoleMember.String(), domain, ResourceCertificate.String(), ActionRead.String()},
		{RoleMember.String(), domain, ResourceDashboard.String(), ActionRead.String()},
		{RoleMember.String(), domain, ResourceAnalytics.String(), ActionRead.String()},
		{RoleMember.String(), domain, ResourceAttendee.String(), ActionRead.String()},
	}
}

// GetAttendeePolicies returns policies for attendee (platform level)
func GetAttendeePolicies() [][]string {
	return [][]string{
		{RoleAttendee.String(), DomainPlatform, ResourceEvent.String(), ActionRead.String()},
		{RoleAttendee.String(), DomainPlatform, ResourceEvent.String(), ActionRegister.String()},
		{RoleAttendee.String(), DomainPlatform, ResourceCertificate.String(), ActionRead.String()},
		{RoleAttendee.String(), DomainPlatform, ResourceCertificate.String(), ActionDownload.String()},
		{RoleAttendee.String(), DomainPlatform, ResourcePayment.String(), ActionRead.String()},
		{RoleAttendee.String(), DomainPlatform, ResourceNotification.String(), ActionRead.String()},
	}
}

// GetPremiumAttendeePolicies returns policies for premium_attendee (platform level)
func GetPremiumAttendeePolicies() [][]string {
	// Premium attendee inherits all attendee policies plus extra
	policies := GetAttendeePolicies()
	
	// Add premium-specific policies
	extraPolicies := [][]string{
		{RolePremiumAttendee.String(), DomainPlatform, ResourceEvent.String(), ActionCreate.String()}, // Can create events? Or other premium features
		{RolePremiumAttendee.String(), DomainPlatform, ResourceDashboard.String(), ActionRead.String()},
		{RolePremiumAttendee.String(), DomainPlatform, ResourceAnalytics.String(), ActionRead.String()},
	}
	
	policies = append(policies, extraPolicies...)
	return policies
}

// GetAllBusinessPolicies returns all business policies for a domain
func GetAllBusinessPolicies(domain string) [][]string {
	var allPolicies [][]string

	allPolicies = append(allPolicies, GetBusinessAdminPolicies(domain)...)
	allPolicies = append(allPolicies, GetEventManagerPolicies(domain)...)
	allPolicies = append(allPolicies, GetMemberPolicies(domain)...)

	return allPolicies
}

// GetBusinessRoleHierarchy returns role inheritance rules for a business domain
func GetBusinessRoleHierarchy(domain string) [][]string {
	return [][]string{
		// Business Admin inherits all business roles
		{RoleBusinessAdmin.String(), RoleEventManager.String(), domain},
		{RoleBusinessAdmin.String(), RoleMember.String(), domain},

		// Event Manager inherits member
		{RoleEventManager.String(), RoleMember.String(), domain},

		// All business roles inherit attendee (platform read-only access)
		{RoleBusinessAdmin.String(), RoleAttendee.String(), DomainPlatform},
		{RoleEventManager.String(), RoleAttendee.String(), DomainPlatform},
		{RoleMember.String(), RoleAttendee.String(), DomainPlatform},
	}
}

// GetPlatformPolicies returns platform-level policies
func GetPlatformPolicies() [][]string {
	var allPolicies [][]string

	// Add attendee policies
	allPolicies = append(allPolicies, GetAttendeePolicies()...)
	
	// Add premium attendee policies
	allPolicies = append(allPolicies, GetPremiumAttendeePolicies()...)

	// Admin policies
	adminPolicies := [][]string{
		{RoleAdmin.String(), DomainPlatform, ResourceBusiness.String(), ActionRead.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceBusiness.String(), ActionUpdate.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceUser.String(), ActionRead.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceUser.String(), ActionUpdate.String()},
		{RoleAdmin.String(), DomainPlatform, ResourcePlatform.String(), ActionManage.String()},
		{RoleAdmin.String(), DomainPlatform, ResourceAnalytics.String(), ActionRead.String()},
	}
	allPolicies = append(allPolicies, adminPolicies...)

	// Super admin policies (full access to everything)
	superAdminPolicies := [][]string{
		{RoleSuperAdmin.String(), DomainPlatform, ResourcePlatform.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceBusiness.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceUser.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourcePayment.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourcePayout.String(), ActionManage.String()},
		{RoleSuperAdmin.String(), DomainPlatform, ResourceAnalytics.String(), ActionManage.String()},
	}
	allPolicies = append(allPolicies, superAdminPolicies...)

	return allPolicies
}

// GetPlatformRoleHierarchy returns role inheritance rules for platform domain
func GetPlatformRoleHierarchy() [][]string {
	return [][]string{
		// Super Admin inherits Admin
		{RoleSuperAdmin.String(), RoleAdmin.String(), DomainPlatform},

		// Admin inherits Premium Attendee
		{RoleAdmin.String(), RolePremiumAttendee.String(), DomainPlatform},

		// Premium Attendee inherits Attendee
		{RolePremiumAttendee.String(), RoleAttendee.String(), DomainPlatform},
	}
}