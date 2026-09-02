// internal/modules/auth/authdomain/constants.go

package authdomain

import "strings"

// Role represents a user role in the system
type Role string

// Context keys for storing values in Fiber context
const (
	ContextKeyUserID        = "user_id"
	ContextKeyUserRole      = "user_role"
	ContextKeyUserEmail     = "user_email"
	ContextKeyUserName      = "user_name"
	ContextKeyDomain        = "domain"
	ContextKeyInstitutionID = "institution_id"
	ContextKeyTeamTypeID    = "team_type_id"
	ContextKeyUserRoles     = "user_roles"
)

const (
	// Platform-level roles
	RoleSuperAdmin Role = "super_admin" // Full platform access
	RoleAdmin      Role = "admin"       // Platform management

	// Team roles (assigned within a team domain)
	RoleAccountAdmin   Role = "account_admin" // Full team management (admin)
	RoleEventManager   Role = "event_manager" // Manage events, attendees, certificates
	RoleTeamMember     Role = "team_member"   // View-only access

	// System
	RoleGuest Role = "guest" // Unregistered user
)

// Resource represents a resource being accessed
type Resource string

const (
	// Platform resources
	ResourcePlatform    Resource = "platform"
	ResourceUser        Resource = "user"
	ResourceInstitution Resource = "institution"
	ResourceTeam        Resource = "team"
	ResourceTeamType    Resource = "team_type"
	
	// Team-level resources (within a team domain)
	ResourceEvent        Resource = "event"
	ResourceCertificate  Resource = "certificate"
	ResourceAttendee     Resource = "attendee"
	ResourcePayment      Resource = "payment"
	ResourcePayout       Resource = "payout"
	ResourceMember       Resource = "member"
	ResourceProfile      Resource = "profile"
	ResourceDashboard    Resource = "dashboard"
	ResourceAnalytics    Resource = "analytics"
	ResourceNotification Resource = "notification"
	ResourceMedia        Resource = "media"
)

// Action represents an operation that can be performed
type Action string

const (
	// Standard actions
	ActionCreate   Action = "create"
	ActionRead     Action = "read"
	ActionUpdate   Action = "update"
	ActionDelete   Action = "delete"
	ActionManage   Action = "manage"   // Full CRUD operations
	ActionIssue    Action = "issue"    // Issue certificates
	ActionRegister Action = "register" // Register for events
	ActionExport   Action = "export"   // Export data (CSV, Excel)
	ActionRefund   Action = "refund"   // Refund payments
	ActionDownload Action = "download" // Download certificates/replays
	ActionInvite   Action = "invite"   // Invite team members

	// ============================================================
	// ✅ NEW: OWN vs ALL Actions for Team Resources
	// ============================================================
	
	// READ - ALL vs OWN
	ActionReadAll  Action = "read_all"  // Read ALL resources in the team
	ActionReadOwn  Action = "read_own"  // Read ONLY OWN resources in the team
	
	// UPDATE - ALL vs OWN
	ActionUpdateAll Action = "update_all" // Update ALL resources in the team
	ActionUpdateOwn Action = "update_own" // Update ONLY OWN resources in the team
	
	// DELETE - ALL vs OWN
	ActionDeleteAll Action = "delete_all" // Delete ALL resources in the team
	ActionDeleteOwn Action = "delete_own" // Delete ONLY OWN resources in the team
	
	// PUBLISH - ALL vs OWN
	ActionPublishAll Action = "publish_all" // Publish ALL resources in the team
	ActionPublishOwn Action = "publish_own" // Publish ONLY OWN resources in the team
)

// Domain constants
const (
	DomainPlatform = "platform"
)

// Team domain prefixes
const (
	TeamDomainPrefixPersonal   = "personal:team:"
	TeamDomainPrefixInstitution = "institution:team:"
)

// ============================================================
// STRING METHODS
// ============================================================

func (r Role) String() string {
	return string(r)
}

func (r Resource) String() string {
	return string(r)
}

func (a Action) String() string {
	return string(a)
}

// ============================================================
// ACTION HELPER METHODS
// ============================================================

// IsOwnAction checks if the action is an "own" action
func (a Action) IsOwnAction() bool {
	switch a {
	case ActionReadOwn, ActionUpdateOwn, ActionDeleteOwn, ActionPublishOwn:
		return true
	default:
		return false
	}
}

// IsAllAction checks if the action is an "all" action
func (a Action) IsAllAction() bool {
	switch a {
	case ActionReadAll, ActionUpdateAll, ActionDeleteAll, ActionPublishAll:
		return true
	default:
		return false
	}
}

// GetOwnAction returns the corresponding "own" action for an action
func (a Action) GetOwnAction() Action {
	switch a {
	case ActionRead:
		return ActionReadOwn
	case ActionUpdate:
		return ActionUpdateOwn
	case ActionDelete:
		return ActionDeleteOwn
	case ActionPublishAll:
		return ActionPublishOwn
	default:
		return a
	}
}

// GetAllAction returns the corresponding "all" action for an action
func (a Action) GetAllAction() Action {
	switch a {
	case ActionRead:
		return ActionReadAll
	case ActionUpdate:
		return ActionUpdateAll
	case ActionDelete:
		return ActionDeleteAll
	case ActionPublishOwn:
		return ActionPublishAll
	default:
		return a
	}
}

// ============================================================
// VALIDATION HELPERS
// ============================================================

func IsValidRole(role string) bool {
	validRoles := map[string]bool{
		RoleSuperAdmin.String():   true,
		RoleAdmin.String():        true,
		RoleAccountAdmin.String(): true,
		RoleEventManager.String(): true,
		RoleTeamMember.String():   true,
		RoleGuest.String():        true,
	}
	return validRoles[role]
}

func IsTeamRole(role string) bool {
	teamRoles := map[string]bool{
		RoleAccountAdmin.String(): true,
		RoleEventManager.String(): true,
		RoleTeamMember.String():   true,
	}
	return teamRoles[role]
}

func IsPlatformRole(role string) bool {
	platformRoles := map[string]bool{
		RoleSuperAdmin.String(): true,
		RoleAdmin.String():      true,
		RoleGuest.String():      true,
	}
	return platformRoles[role]
}

func IsValidTeamRole(role string) bool {
	return IsTeamRole(role)
}

// ============================================================
// DOMAIN HELPERS
// ============================================================

// PersonalTeamDomain returns the personal team domain for a user
// Format: "personal:team:{user_id}"
func PersonalTeamDomain(userID string) string {
	if userID == "" {
		return ""
	}
	return TeamDomainPrefixPersonal + userID
}

// InstitutionTeamDomain returns the institution team domain
// Format: "institution:team:{institution_id}"
func InstitutionTeamDomain(institutionID string) string {
	if institutionID == "" {
		return ""
	}
	return TeamDomainPrefixInstitution + institutionID
}

// IsPersonalTeamDomain checks if a domain is a personal team domain
func IsPersonalTeamDomain(domain string) bool {
	return strings.HasPrefix(domain, TeamDomainPrefixPersonal)
}

// IsInstitutionTeamDomain checks if a domain is an institution team domain
func IsInstitutionTeamDomain(domain string) bool {
	return strings.HasPrefix(domain, TeamDomainPrefixInstitution)
}

// IsTeamDomain checks if a domain is a team domain (personal or institution)
func IsTeamDomain(domain string) bool {
	return IsPersonalTeamDomain(domain) || IsInstitutionTeamDomain(domain)
}

// IsPlatformDomain checks if a domain is the platform domain
func IsPlatformDomain(domain string) bool {
	return domain == DomainPlatform
}

// ExtractTeamID extracts team ID from a team domain
func ExtractTeamID(domain string) string {
	if IsPersonalTeamDomain(domain) {
		return strings.TrimPrefix(domain, TeamDomainPrefixPersonal)
	}
	if IsInstitutionTeamDomain(domain) {
		return strings.TrimPrefix(domain, TeamDomainPrefixInstitution)
	}
	return ""
}

// ExtractTeamType extracts team type from a team domain
// Returns: "personal", "institution", or empty string
func ExtractTeamType(domain string) string {
	if IsPersonalTeamDomain(domain) {
		return "personal"
	}
	if IsInstitutionTeamDomain(domain) {
		return "institution"
	}
	return ""
}

// ============================================================
// GET ALL HELPERS
// ============================================================

func GetAllRoles() []Role {
	return []Role{
		RoleSuperAdmin,
		RoleAdmin,
		RoleAccountAdmin,
		RoleEventManager,
		RoleTeamMember,
		RoleGuest,
	}
}

func GetAllTeamRoles() []Role {
	return []Role{
		RoleAccountAdmin,
		RoleEventManager,
		RoleTeamMember,
	}
}

func GetAllPlatformRoles() []Role {
	return []Role{
		RoleSuperAdmin,
		RoleAdmin,
		RoleGuest,
	}
}

func GetAllResources() []Resource {
	return []Resource{
		ResourcePlatform,
		ResourceUser,
		ResourceInstitution,
		ResourceTeam,
		ResourceTeamType,
		ResourceEvent,
		ResourceCertificate,
		ResourceAttendee,
		ResourcePayment,
		ResourcePayout,
		ResourceMember,
		ResourceProfile,
		ResourceDashboard,
		ResourceAnalytics,
		ResourceNotification,
		ResourceMedia,
	}
}

func GetAllTeamResources() []Resource {
	return []Resource{
		ResourceEvent,
		ResourceCertificate,
		ResourceAttendee,
		ResourcePayment,
		ResourcePayout,
		ResourceMember,
		ResourceProfile,
		ResourceDashboard,
		ResourceAnalytics,
		ResourceNotification,
		ResourceMedia,
	}
}

func GetAllPlatformResources() []Resource {
	return []Resource{
		ResourcePlatform,
		ResourceUser,
		ResourceInstitution,
		ResourceTeam,
		ResourceTeamType,
	}
}

func GetAllActions() []Action {
	return []Action{
		ActionCreate,
		ActionRead,
		ActionUpdate,
		ActionDelete,
		ActionManage,
		ActionIssue,
		ActionRegister,
		ActionExport,
		ActionRefund,
		ActionDownload,
		ActionInvite,
		ActionReadAll,
		ActionReadOwn,
		ActionUpdateAll,
		ActionUpdateOwn,
		ActionDeleteAll,
		ActionDeleteOwn,
		ActionPublishAll,
		ActionPublishOwn,
	}
}

// ============================================================
// PERMISSION MATRIX HELPERS
// ============================================================

func DefaultPlatformPermissions() map[Role][]string {
	return map[Role][]string{
		RoleSuperAdmin: {
			"*:*",
		},
		RoleAdmin: {
			"user:read",
			"user:update",
			"user:delete",
			"institution:read",
			"institution:update",
			"institution:delete",
			"event:read_all",
			"event:update_all",
			"event:delete_all",
			"team:read",
			"team:update",
			"team:delete",
			"analytics:read",
			"payment:read",
			"certificate:read",
			"member:read",
			"member:delete",
		},
	}
}

func DefaultTeamPermissions() map[Role][]string {
	return map[Role][]string{
		RoleAccountAdmin: {
			"event:create",
			"event:read_all",
			"event:update_all",
			"event:delete_all",
			"event:publish_all",
			"event:manage",
			"certificate:create",
			"certificate:read",
			"certificate:update",
			"certificate:issue",
			"certificate:delete",
			"attendee:read",
			"attendee:update",
			"attendee:export",
			"payment:read",
			"payment:create",
			"payment:refund",
			"member:create",
			"member:read",
			"member:update",
			"member:delete",
			"member:invite",
			"institution:read",
			"institution:update",
			"team:create",
			"team:read",
			"team:update",
			"team:delete",
			"dashboard:read",
			"profile:read",
			"profile:update",
		},
		RoleEventManager: {
			"event:create",
			"event:read_all",
			"event:update_all",
			"event:delete_all",
			"event:publish_all",
			"attendee:read",
			"attendee:update",
			"attendee:export",
			"certificate:create",
			"certificate:read",
			"certificate:issue",
			"certificate:delete",
			"payment:read",
			"dashboard:read",
			"team:read",
		},
		RoleTeamMember: {
			"event:read_own",
			"event:update_own",
			"event:delete_own",
			"event:publish_own",
			"attendee:read",
			"attendee:update_own",
			"certificate:read",
			"dashboard:read",
			"team:read",
		},
	}
}

// ============================================================
// ROLE PRIORITY HELPERS
// ============================================================

func RolePriority(role Role) int {
	priority := map[Role]int{
		RoleSuperAdmin:   100,
		RoleAdmin:        90,
		RoleAccountAdmin: 80,
		RoleEventManager: 70,
		RoleTeamMember:   60,
		RoleGuest:        10,
	}
	return priority[role]
}

func HasHigherOrEqualPriority(role1, role2 Role) bool {
	return RolePriority(role1) >= RolePriority(role2)
}

func IsRoleAtLeast(role Role, minRole Role) bool {
	return RolePriority(role) >= RolePriority(minRole)
}