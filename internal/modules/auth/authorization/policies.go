// internal/modules/auth/authorization/policies.go

package authorization

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"

// ============================================================
// PERSONAL TEAM POLICIES
// ============================================================

// GetPersonalTeamPolicies returns all policies for a personal team
// domain should be "personal:team:{user_id}"
func GetPersonalTeamPolicies(domain string) [][]string {
	role := authdomain.RoleAccountAdmin.String()

	return [][]string{
		// Event permissions
		{role, domain, "event", "create"},
		{role, domain, "event", "read"},
		{role, domain, "event", "update"},
		{role, domain, "event", "delete"},
		{role, domain, "event", "manage"},

		// Certificate permissions
		{role, domain, "certificate", "create"},
		{role, domain, "certificate", "read"},
		{role, domain, "certificate", "update"},
		{role, domain, "certificate", "issue"},
		{role, domain, "certificate", "delete"},

		// Attendee permissions
		{role, domain, "attendee", "read"},
		{role, domain, "attendee", "update"},
		{role, domain, "attendee", "export"},

		// Payment permissions
		{role, domain, "payment", "read"},
		{role, domain, "payment", "create"},
		{role, domain, "payment", "refund"},

		// Member permissions
		{role, domain, "member", "create"},
		{role, domain, "member", "read"},
		{role, domain, "member", "update"},
		{role, domain, "member", "delete"},
		{role, domain, "member", "invite"},

		// Institution permissions (limited for personal)
		{role, domain, "institution", "read"},

		// Team permissions
		{role, domain, "team", "read"},
		{role, domain, "team", "update"},

		// Dashboard and profile
		{role, domain, "dashboard", "read"},
		{role, domain, "profile", "read"},
		{role, domain, "profile", "update"},
	}
}

// ============================================================
// INSTITUTION TEAM POLICIES
// ============================================================

// GetInstitutionTeamPolicies returns all policies for an institution team
// domain should be "institution:team:{institution_id}"
func GetInstitutionTeamPolicies(domain string) [][]string {
	role := authdomain.RoleAccountAdmin.String()

	return [][]string{
		// Event permissions
		{role, domain, "event", "create"},
		{role, domain, "event", "read"},
		{role, domain, "event", "update"},
		{role, domain, "event", "delete"},
		{role, domain, "event", "manage"},

		// Certificate permissions
		{role, domain, "certificate", "create"},
		{role, domain, "certificate", "read"},
		{role, domain, "certificate", "update"},
		{role, domain, "certificate", "issue"},
		{role, domain, "certificate", "delete"},

		// Attendee permissions
		{role, domain, "attendee", "read"},
		{role, domain, "attendee", "update"},
		{role, domain, "attendee", "export"},

		// Payment permissions
		{role, domain, "payment", "read"},
		{role, domain, "payment", "create"},
		{role, domain, "payment", "refund"},

		// Member permissions
		{role, domain, "member", "create"},
		{role, domain, "member", "read"},
		{role, domain, "member", "update"},
		{role, domain, "member", "delete"},
		{role, domain, "member", "invite"},

		// Institution permissions (full for institution)
		{role, domain, "institution", "read"},
		{role, domain, "institution", "update"},
		{role, domain, "institution", "manage"},

		// Team permissions
		{role, domain, "team", "create"},
		{role, domain, "team", "read"},
		{role, domain, "team", "update"},
		{role, domain, "team", "delete"},

		// Dashboard and profile
		{role, domain, "dashboard", "read"},
		{role, domain, "profile", "read"},
		{role, domain, "profile", "update"},
	}
}

// ============================================================
// TEAM ROLE HIERARCHY
// ============================================================

// GetTeamRoleHierarchy returns the role hierarchy for a team domain
// domain should be "personal:team:{id}" or "institution:team:{id}"
func GetTeamRoleHierarchy(domain string) [][]string {
	return [][]string{
		// Account Admin inherits Event Manager and Team Member
		{authdomain.RoleAccountAdmin.String(), authdomain.RoleEventManager.String(), domain},
		{authdomain.RoleAccountAdmin.String(), authdomain.RoleTeamMember.String(), domain},

		// Event Manager inherits Team Member
		{authdomain.RoleEventManager.String(), authdomain.RoleTeamMember.String(), domain},
	}
}

// ============================================================
// PLATFORM POLICIES
// ============================================================

// GetPlatformPolicies returns platform-level policies
func GetPlatformPolicies() [][]string {
	var allPolicies [][]string

	// Admin policies
	adminPolicies := [][]string{
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceUser.String(), authdomain.ActionRead.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceUser.String(), authdomain.ActionUpdate.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceUser.String(), authdomain.ActionDelete.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceInstitution.String(), authdomain.ActionRead.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceInstitution.String(), authdomain.ActionUpdate.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceInstitution.String(), authdomain.ActionDelete.String()},
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
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceTeam.String(), authdomain.ActionUpdate.String()},
		{authdomain.RoleAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceTeam.String(), authdomain.ActionDelete.String()},
	}
	allPolicies = append(allPolicies, adminPolicies...)

	// Super admin policies (full access to everything)
	superAdminPolicies := [][]string{
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourcePlatform.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceUser.String(), authdomain.ActionManage.String()},
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
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceTeam.String(), authdomain.ActionManage.String()},
		{authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform, authdomain.ResourceTeamType.String(), authdomain.ActionManage.String()},
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

// ============================================================
// EVENT MANAGER POLICIES
// ============================================================

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

		// Team - Read only
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},
	}
}

// ============================================================
// TEAM MEMBER POLICIES (View-only)
// ============================================================

// GetTeamMemberPolicies returns policies for team_member (view-only)
func GetTeamMemberPolicies(domain string) [][]string {
	return [][]string{
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionRead.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceAnalytics.String(), authdomain.ActionRead.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceAttendee.String(), authdomain.ActionRead.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceMedia.String(), authdomain.ActionRead.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},
	}
}