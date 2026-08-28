// internal/modules/auth/authdomain/permission_service.go

package authdomain

import "context"

// ============================================================
// OUTBOUND PORT: PermissionService
// Defines what the auth module NEEDS for authorization
// ============================================================

type PermissionService interface {
	// ============================================================
	// PERSONAL TEAM ROLE ASSIGNMENT
	// Domain: personal:team:{user_id}
	// ============================================================

	// AssignPersonalTeamAdmin assigns account_admin role to a user for their personal team
	AssignPersonalTeamAdmin(ctx context.Context, userID string) error

	// AssignEventManagerRoleForPersonalTeam assigns event_manager role to a user for their personal team
	AssignEventManagerRoleForPersonalTeam(ctx context.Context, teamOwnerID, userID string) error

	// AssignTeamMemberRoleForPersonalTeam assigns team_member role to a user for their personal team
	AssignTeamMemberRoleForPersonalTeam(ctx context.Context, teamOwnerID, userID string) error

	// ============================================================
	// INSTITUTION TEAM ROLE ASSIGNMENT
	// Domain: institution:team:{institution_id}
	// ============================================================

	// AssignInstitutionAdminRole assigns account_admin role to a user for an institution team
	AssignInstitutionAdminRole(ctx context.Context, institutionID, userID string) error

	// AssignEventManagerRoleForInstitution assigns event_manager role to a user for an institution team
	AssignEventManagerRoleForInstitution(ctx context.Context, institutionID, userID string) error

	// AssignTeamMemberRoleForInstitution assigns team_member role to a user for an institution team
	AssignTeamMemberRoleForInstitution(ctx context.Context, institutionID, userID string) error

	// ============================================================
	// TEAM POLICY MANAGEMENT
	// ============================================================

	// AddPersonalTeamPolicies adds default policies for a personal team
	// Domain: personal:team:{user_id}
	AddPersonalTeamPolicies(ctx context.Context, userID string) error

	// AddInstitutionPolicies adds default policies for an institution team
	// Domain: institution:team:{institution_id}
	AddInstitutionPolicies(ctx context.Context, institutionID string) error

	// RemovePersonalTeamPolicies removes all policies for a personal team
	// Domain: personal:team:{user_id}
	RemovePersonalTeamPolicies(ctx context.Context, userID string) error

	// RemoveInstitutionTeamPolicies removes all policies for an institution team
	// Domain: institution:team:{institution_id}
	RemoveInstitutionTeamPolicies(ctx context.Context, institutionID string) error

	// ============================================================
	// PLATFORM ROLE METHODS
	// Domain: platform
	// ============================================================

	// AssignAdminRole assigns platform admin role to a user
	AssignAdminRole(ctx context.Context, userID string) error

	// AssignSuperAdminRole assigns super admin role to a user
	AssignSuperAdminRole(ctx context.Context, userID string) error

	// RemovePlatformRole removes a platform role from a user
	RemovePlatformRole(ctx context.Context, userID string, role string) error

	// AddPlatformPolicies adds default platform policies
	AddPlatformPolicies(ctx context.Context) error

	// ============================================================
	// ROLE REMOVAL METHODS
	// ============================================================

	// RemovePersonalTeamRole removes a specific role from a user in a personal team
	// Domain: personal:team:{teamOwnerID}
	RemovePersonalTeamRole(ctx context.Context, teamOwnerID, userID string, role string) error

	// RemoveInstitutionTeamRole removes a specific role from a user in an institution team
	// Domain: institution:team:{institutionID}
	RemoveInstitutionTeamRole(ctx context.Context, institutionID, userID string, role string) error

	// RemoveAllPersonalTeamRoles removes all roles for a user from a personal team
	// Domain: personal:team:{teamOwnerID}
	RemoveAllPersonalTeamRoles(ctx context.Context, teamOwnerID, userID string) error

	// RemoveAllInstitutionTeamRoles removes all roles for a user from an institution team
	// Domain: institution:team:{institutionID}
	RemoveAllInstitutionTeamRoles(ctx context.Context, institutionID, userID string) error

	// ============================================================
	// PERSONAL TEAM PERMISSION CHECKS
	// Domain: personal:team:{teamOwnerID}
	// ============================================================

	// HasPersonalTeamPermission checks if a user has permission in a personal team
	HasPersonalTeamPermission(ctx context.Context, userID, teamOwnerID, resource, action string) bool

	// ============================================================
	// INSTITUTION TEAM PERMISSION CHECKS
	// Domain: institution:team:{institutionID}
	// ============================================================

	// HasInstitutionTeamPermission checks if a user has permission in an institution team
	HasInstitutionTeamPermission(ctx context.Context, userID, institutionID, resource, action string) bool

	// ============================================================
	// GENERIC TEAM PERMISSION CHECKS (Try both domains)
	// ============================================================

	// CanManageTeamEvent checks if a user can manage events in a team
	CanManageTeamEvent(ctx context.Context, teamID, userID string) bool

	// CanIssueTeamCertificate checks if a user can issue certificates in a team
	CanIssueTeamCertificate(ctx context.Context, teamID, userID string) bool

	// CanManageTeam checks if a user can manage a team
	CanManageTeam(ctx context.Context, teamID, userID string) bool

	// ============================================================
	// PERSONAL TEAM ROLE CHECKS
	// Domain: personal:team:{teamOwnerID}
	// ============================================================

	// IsPersonalTeamAdmin checks if a user is an admin of a personal team
	IsPersonalTeamAdmin(ctx context.Context, userID, teamOwnerID string) bool

	// IsPersonalTeamEventManager checks if a user is an event manager of a personal team
	IsPersonalTeamEventManager(ctx context.Context, userID, teamOwnerID string) bool

	// IsPersonalTeamMember checks if a user is a member of a personal team
	IsPersonalTeamMember(ctx context.Context, userID, teamOwnerID string) bool

	// ============================================================
	// INSTITUTION TEAM ROLE CHECKS
	// Domain: institution:team:{institutionID}
	// ============================================================

	// IsInstitutionTeamAdmin checks if a user is an admin of an institution team
	IsInstitutionTeamAdmin(ctx context.Context, userID, institutionID string) bool

	// IsInstitutionTeamEventManager checks if a user is an event manager of an institution team
	IsInstitutionTeamEventManager(ctx context.Context, userID, institutionID string) bool

	// IsInstitutionTeamMember checks if a user is a member of an institution team
	IsInstitutionTeamMember(ctx context.Context, userID, institutionID string) bool

	// ============================================================
	// USER INFORMATION METHODS
	// ============================================================

	// GetUserRoles returns all roles for a user in a domain
	GetUserRoles(ctx context.Context, userID, domain string) ([]string, error)

	// GetUserTeamIDs returns all team IDs where a user has roles (both personal and institution)
	GetUserTeamIDs(ctx context.Context, userID string) []string

	// GetUserPersonalTeamIDs returns personal team IDs where a user has roles
	GetUserPersonalTeamIDs(ctx context.Context, userID string) []string

	// GetUserInstitutionTeamIDs returns institution team IDs where a user has roles
	GetUserInstitutionTeamIDs(ctx context.Context, userID string) []string

	// HasTeamAccess checks if a user has any team role
	HasTeamAccess(ctx context.Context, userID string) bool

	// GetRolesForUser returns all roles for a user in a domain (alias for GetUserRoles)
	GetRolesForUser(ctx context.Context, userID, domain string) ([]string, error)
}