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
	eventManager := authdomain.RoleEventManager.String()
	teamMember := authdomain.RoleTeamMember.String()

	var policies [][]string

	// ============================================================
	// ACCOUNT ADMIN - Full access (ALL)
	// ============================================================
	accountAdminPolicies := [][]string{
		// Event permissions - ALL
		{accountAdmin, domain, "event", "create"},
		{accountAdmin, domain, "event", "read_all"},
		{accountAdmin, domain, "event", "update_all"},
		{accountAdmin, domain, "event", "delete_all"},
		{accountAdmin, domain, "event", "publish_all"},
		{accountAdmin, domain, "event", "manage"},

		// Certificate permissions
		{accountAdmin, domain, "certificate", "create"},
		{accountAdmin, domain, "certificate", "read"},
		{accountAdmin, domain, "certificate", "update"},
		{accountAdmin, domain, "certificate", "issue"},
		{accountAdmin, domain, "certificate", "delete"},

		// Attendee permissions
		{accountAdmin, domain, "attendee", "read"},
		{accountAdmin, domain, "attendee", "update"},
		{accountAdmin, domain, "attendee", "export"},

		// Payment permissions
		{accountAdmin, domain, "payment", "read"},
		{accountAdmin, domain, "payment", "create"},
		{accountAdmin, domain, "payment", "refund"},

		// Member permissions
		{accountAdmin, domain, "member", "create"},
		{accountAdmin, domain, "member", "read"},
		{accountAdmin, domain, "member", "update"},
		{accountAdmin, domain, "member", "delete"},
		{accountAdmin, domain, "member", "invite"},

		// Institution permissions (limited for personal)
		{accountAdmin, domain, "institution", "read"},

		// Team permissions
		{accountAdmin, domain, "team", "read"},
		{accountAdmin, domain, "team", "update"},
		{accountAdmin, domain, "team", "delete"},

		// Dashboard and profile
		{accountAdmin, domain, "dashboard", "read"},
		{accountAdmin, domain, "profile", "read"},
		{accountAdmin, domain, "profile", "update"},
	}
	policies = append(policies, accountAdminPolicies...)

	// ============================================================
	// EVENT MANAGER - Full event management (ALL)
	// ============================================================
	eventManagerPolicies := [][]string{
		// Event permissions - ALL
		{eventManager, domain, "event", "create"},
		{eventManager, domain, "event", "read_all"},
		{eventManager, domain, "event", "update_all"},
		{eventManager, domain, "event", "delete_all"},
		{eventManager, domain, "event", "publish_all"},
		{eventManager, domain, "event", "manage"},

		// Attendee permissions
		{eventManager, domain, "attendee", "read"},
		{eventManager, domain, "attendee", "update"},
		{eventManager, domain, "attendee", "export"},

		// Certificate permissions
		{eventManager, domain, "certificate", "create"},
		{eventManager, domain, "certificate", "read"},
		{eventManager, domain, "certificate", "issue"},
		{eventManager, domain, "certificate", "delete"},

		// Payment - Read only
		{eventManager, domain, "payment", "read"},

		// Dashboard
		{eventManager, domain, "dashboard", "read"},

		// Team - Read only
		{eventManager, domain, "team", "read"},
	}
	policies = append(policies, eventManagerPolicies...)

	// ============================================================
	// TEAM MEMBER - Own access only (OWN)
	// ============================================================
	teamMemberPolicies := [][]string{
		// Event permissions - OWN only
		{teamMember, domain, "event", "read_own"},
		{teamMember, domain, "event", "update_own"},
		{teamMember, domain, "event", "delete_own"},
		{teamMember, domain, "event", "publish_own"},

		// Attendee permissions
		{teamMember, domain, "attendee", "read"},
		{teamMember, domain, "attendee", "update_own"},

		// Certificate - Read only
		{teamMember, domain, "certificate", "read"},

		// Dashboard
		{teamMember, domain, "dashboard", "read"},

		// Team - Read only
		{teamMember, domain, "team", "read"},
	}
	policies = append(policies, teamMemberPolicies...)

	return policies
}

// ============================================================
// INSTITUTION TEAM POLICIES
// ============================================================

// GetInstitutionTeamPolicies returns all policies for an institution team
// domain should be "institution:team:{institution_id}"
func GetInstitutionTeamPolicies(domain string) [][]string {
	accountAdmin := authdomain.RoleAccountAdmin.String()
	eventManager := authdomain.RoleEventManager.String()
	teamMember := authdomain.RoleTeamMember.String()

	var policies [][]string

	// ============================================================
	// ACCOUNT ADMIN - Full access (ALL)
	// ============================================================
	accountAdminPolicies := [][]string{
		// Event permissions - ALL
		{accountAdmin, domain, "event", "create"},
		{accountAdmin, domain, "event", "read_all"},
		{accountAdmin, domain, "event", "update_all"},
		{accountAdmin, domain, "event", "delete_all"},
		{accountAdmin, domain, "event", "publish_all"},
		{accountAdmin, domain, "event", "manage"},

		// Certificate permissions
		{accountAdmin, domain, "certificate", "create"},
		{accountAdmin, domain, "certificate", "read"},
		{accountAdmin, domain, "certificate", "update"},
		{accountAdmin, domain, "certificate", "issue"},
		{accountAdmin, domain, "certificate", "delete"},

		// Attendee permissions
		{accountAdmin, domain, "attendee", "read"},
		{accountAdmin, domain, "attendee", "update"},
		{accountAdmin, domain, "attendee", "export"},

		// Payment permissions
		{accountAdmin, domain, "payment", "read"},
		{accountAdmin, domain, "payment", "create"},
		{accountAdmin, domain, "payment", "refund"},

		// Member permissions
		{accountAdmin, domain, "member", "create"},
		{accountAdmin, domain, "member", "read"},
		{accountAdmin, domain, "member", "update"},
		{accountAdmin, domain, "member", "delete"},
		{accountAdmin, domain, "member", "invite"},

		// Institution permissions (full for institution)
		{accountAdmin, domain, "institution", "read"},
		{accountAdmin, domain, "institution", "update"},
		{accountAdmin, domain, "institution", "manage"},

		// Team permissions
		{accountAdmin, domain, "team", "create"},
		{accountAdmin, domain, "team", "read"},
		{accountAdmin, domain, "team", "update"},
		{accountAdmin, domain, "team", "delete"},

		// Dashboard and profile
		{accountAdmin, domain, "dashboard", "read"},
		{accountAdmin, domain, "profile", "read"},
		{accountAdmin, domain, "profile", "update"},
	}
	policies = append(policies, accountAdminPolicies...)

	// ============================================================
	// EVENT MANAGER - Full event management (ALL)
	// ============================================================
	eventManagerPolicies := [][]string{
		// Event permissions - ALL
		{eventManager, domain, "event", "create"},
		{eventManager, domain, "event", "read_all"},
		{eventManager, domain, "event", "update_all"},
		{eventManager, domain, "event", "delete_all"},
		{eventManager, domain, "event", "publish_all"},
		{eventManager, domain, "event", "manage"},

		// Attendee permissions
		{eventManager, domain, "attendee", "read"},
		{eventManager, domain, "attendee", "update"},
		{eventManager, domain, "attendee", "export"},

		// Certificate permissions
		{eventManager, domain, "certificate", "create"},
		{eventManager, domain, "certificate", "read"},
		{eventManager, domain, "certificate", "issue"},
		{eventManager, domain, "certificate", "delete"},

		// Payment - Read only
		{eventManager, domain, "payment", "read"},

		// Dashboard
		{eventManager, domain, "dashboard", "read"},

		// Team - Read only
		{eventManager, domain, "team", "read"},
	}
	policies = append(policies, eventManagerPolicies...)

	// ============================================================
	// TEAM MEMBER - Own access only (OWN)
	// ============================================================
	teamMemberPolicies := [][]string{
		// Event permissions - OWN only
		{teamMember, domain, "event", "read_own"},
		{teamMember, domain, "event", "update_own"},
		{teamMember, domain, "event", "delete_own"},
		{teamMember, domain, "event", "publish_own"},

		// Attendee permissions
		{teamMember, domain, "attendee", "read"},
		{teamMember, domain, "attendee", "update_own"},

		// Certificate - Read only
		{teamMember, domain, "certificate", "read"},

		// Dashboard
		{teamMember, domain, "dashboard", "read"},

		// Team - Read only
		{teamMember, domain, "team", "read"},
	}
	policies = append(policies, teamMemberPolicies...)

	return policies
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
		// Event management - ALL
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdateAll.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionDeleteAll.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionPublishAll.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionManage.String()},

		// Certificate management
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceCertificate.String(), authdomain.ActionCreate.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String()},
		{authdomain.RoleEventManager.String(), domain, authdomain.ResourceCertificate.String(), authdomain.ActionDelete.String()},

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
// TEAM MEMBER POLICIES (OWN access only)
// ============================================================

// GetTeamMemberPolicies returns policies for team_member (OWN access only)
func GetTeamMemberPolicies(domain string) [][]string {
	return [][]string{
		// Event permissions - OWN only
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionReadOwn.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdateOwn.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionDeleteOwn.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceEvent.String(), authdomain.ActionPublishOwn.String()},

		// Attendee - Read and update own
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceAttendee.String(), authdomain.ActionRead.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceAttendee.String(), authdomain.ActionUpdateOwn.String()},

		// Certificate - Read only
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceCertificate.String(), authdomain.ActionRead.String()},

		// Dashboard - Read only
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceDashboard.String(), authdomain.ActionRead.String()},
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceAnalytics.String(), authdomain.ActionRead.String()},

		// Media - Read only
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceMedia.String(), authdomain.ActionRead.String()},

		// Team - Read only
		{authdomain.RoleTeamMember.String(), domain, authdomain.ResourceTeam.String(), authdomain.ActionRead.String()},
	}
}