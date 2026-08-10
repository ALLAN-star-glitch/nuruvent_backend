// internal/modules/authorization/constants.go

package authorization

// Role represents a user role in the system
type Role string

// Context keys for storing values in Fiber context
const (
	ContextKeyUserID       = "user_id"
	ContextKeyUserRole     = "user_role"
	ContextKeyUserEmail    = "user_email"
	ContextKeyUserName     = "user_name"
	ContextKeyDomain       = "domain"
	ContextKeyAccountID    = "account_id"
	ContextKeyInstitutionID = "institution_id"
	ContextKeyUserRoles    = "user_roles"
)

const (
	// Platform-level roles
	RoleSuperAdmin Role = "super_admin" // Full platform access
	RoleAdmin      Role = "admin"       // Platform management

	// Account team roles (assigned within an account domain)
	RoleAccountAdmin   Role = "account_admin"   // Full account management
	RoleEventManager   Role = "event_manager"   // Manage events, attendees, certificates
	RoleTeamMember     Role = "team_member"     // View-only access

	// System
	RoleGuest Role = "guest" // Unregistered user
)

// Resource represents a resource being accessed
type Resource string

const (
	ResourceEvent        Resource = "event"
	ResourceCertificate  Resource = "certificate"
	ResourceAttendee     Resource = "attendee"
	ResourcePayment      Resource = "payment"
	ResourcePayout       Resource = "payout"
	ResourceInstitution  Resource = "institution"
	ResourceMember       Resource = "member"
	ResourceAccount      Resource = "account"
	ResourceProfile      Resource = "profile"
	ResourceDashboard    Resource = "dashboard"
	ResourceAnalytics    Resource = "analytics"
	ResourceNotification Resource = "notification"
	ResourcePlatform     Resource = "platform"
	ResourceMedia        Resource = "media"
)

// Action represents an operation that can be performed
type Action string

const (
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
)

// Domain constants
const (
	DomainPlatform = "platform"
)

// ============================================================
// STRING METHODS
// ============================================================

// String returns the string representation of a Role
func (r Role) String() string {
	return string(r)
}

// String returns the string representation of a Resource
func (r Resource) String() string {
	return string(r)
}

// String returns the string representation of an Action
func (a Action) String() string {
	return string(a)
}

// ============================================================
// VALIDATION HELPERS
// ============================================================

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	validRoles := map[string]bool{
		RoleSuperAdmin.String():  true,
		RoleAdmin.String():       true,
		RoleAccountAdmin.String(): true,
		RoleEventManager.String(): true,
		RoleTeamMember.String():  true,
		RoleGuest.String():       true,
	}
	return validRoles[role]
}

// IsTeamRole checks if a role is a team-level role (within an account domain)
func IsTeamRole(role string) bool {
	teamRoles := map[string]bool{
		RoleAccountAdmin.String():  true,
		RoleEventManager.String(): true,
		RoleTeamMember.String():   true,
	}
	return teamRoles[role]
}

// IsPlatformRole checks if a role is a platform-level role
func IsPlatformRole(role string) bool {
	platformRoles := map[string]bool{
		RoleSuperAdmin.String(): true,
		RoleAdmin.String():      true,
		RoleGuest.String():      true,
	}
	return platformRoles[role]
}

// ============================================================
// DOMAIN HELPERS
// ============================================================

// AccountDomain returns the account domain string
func AccountDomain(accountID string) string {
	if accountID == "" {
		return ""
	}
	return "account:" + accountID
}

// IsAccountDomain checks if a domain is an account domain
func IsAccountDomain(domain string) bool {
	return len(domain) > 8 && domain[:8] == "account:"
}

// IsPlatformDomain checks if a domain is the platform domain
func IsPlatformDomain(domain string) bool {
	return domain == DomainPlatform
}

// ExtractAccountID extracts account ID from domain
func ExtractAccountID(domain string) string {
	if IsAccountDomain(domain) {
		return domain[8:]
	}
	return ""
}

// ============================================================
// GET ALL HELPERS
// ============================================================

// GetAllRoles returns all defined roles
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

// GetAllTeamRoles returns all team-level roles
func GetAllTeamRoles() []Role {
	return []Role{
		RoleAccountAdmin,
		RoleEventManager,
		RoleTeamMember,
	}
}

// GetAllPlatformRoles returns all platform-level roles
func GetAllPlatformRoles() []Role {
	return []Role{
		RoleSuperAdmin,
		RoleAdmin,
		RoleGuest,
	}
}

// GetAllResources returns all defined resources
func GetAllResources() []Resource {
	return []Resource{
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
		ResourcePlatform,
		ResourceMedia,
	}
}

// GetAllActions returns all defined actions
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
	}
}