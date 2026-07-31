package authorization

// Role represents a user role in the system
type Role string

// Context keys for storing values in Fiber context
const (
	ContextKeyUserID      = "user_id"
	ContextKeyUserRole    = "user_role"
	ContextKeyUserEmail   = "user_email"
	ContextKeyUserName    = "user_name"
	ContextKeyDomain      = "domain"
	ContextKeyBusinessID  = "business_id"
	ContextKeyUserRoles   = "user_roles"
)

const (
	// Platform-level roles
	RoleSuperAdmin Role = "super_admin" // Full platform access
	RoleAdmin      Role = "admin"       // Platform management

	// Business roles (assigned within a business domain)
	RoleBusinessAdmin   Role = "business_admin"   // Full business management (was "host")
	RoleEventManager    Role = "event_manager"    // Manage events, attendees, certificates
	RoleMember          Role = "member"           // View-only access to business data

	// Consumer roles (platform level)
	RoleAttendee        Role = "attendee"         // Event participant
	RolePremiumAttendee Role = "premium_attendee" // Premium event participant

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
	ResourceBusiness     Resource = "business"
	ResourceMember       Resource = "member"
	ResourceUser         Resource = "user"
	ResourceDashboard    Resource = "dashboard"
	ResourceAnalytics    Resource = "analytics"
	ResourceNotification Resource = "notification"
	ResourcePlatform     Resource = "platform"
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
		RoleSuperAdmin.String():     true,
		RoleAdmin.String():          true,
		RoleBusinessAdmin.String():  true,
		RoleEventManager.String():   true,
		RoleMember.String():         true,
		RoleAttendee.String():       true,
		RolePremiumAttendee.String(): true,
		RoleGuest.String():          true,
	}
	return validRoles[role]
}

// IsBusinessRole checks if a role is a business-level role
func IsBusinessRole(role string) bool {
	businessRoles := map[string]bool{
		RoleBusinessAdmin.String():  true,
		RoleEventManager.String():   true,
		RoleMember.String():         true,
	}
	return businessRoles[role]
}

// IsPlatformRole checks if a role is a platform-level role
func IsPlatformRole(role string) bool {
	platformRoles := map[string]bool{
		RoleSuperAdmin.String(): true,
		RoleAdmin.String():      true,
		RoleAttendee.String():   true,
		RolePremiumAttendee.String(): true,
		RoleGuest.String():      true,
	}
	return platformRoles[role]
}

// ============================================================
// DOMAIN HELPERS
// ============================================================

// BusinessDomain returns the business domain string
func BusinessDomain(businessID string) string {
	if businessID == "" {
		return ""
	}
	return "business:" + businessID
}

// UserDomain returns the user domain string
func UserDomain(userID string) string {
	if userID == "" {
		return ""
	}
	return "user:" + userID
}

// IsBusinessDomain checks if a domain is a business domain
func IsBusinessDomain(domain string) bool {
	return len(domain) > 9 && domain[:9] == "business:"
}

// IsUserDomain checks if a domain is a user domain
func IsUserDomain(domain string) bool {
	return len(domain) > 5 && domain[:5] == "user:"
}

// IsPlatformDomain checks if a domain is the platform domain
func IsPlatformDomain(domain string) bool {
	return domain == DomainPlatform
}

// ExtractBusinessID extracts business ID from domain
func ExtractBusinessID(domain string) string {
	if IsBusinessDomain(domain) {
		return domain[9:]
	}
	return ""
}

// ExtractUserID extracts user ID from domain
func ExtractUserID(domain string) string {
	if IsUserDomain(domain) {
		return domain[5:]
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
		RoleBusinessAdmin,
		RoleEventManager,
		RoleMember,
		RoleAttendee,
		RolePremiumAttendee,
		RoleGuest,
	}
}

// GetAllBusinessRoles returns all business-level roles
func GetAllBusinessRoles() []Role {
	return []Role{
		RoleBusinessAdmin,
		RoleEventManager,
		RoleMember,
	}
}

// GetAllPlatformRoles returns all platform-level roles
func GetAllPlatformRoles() []Role {
	return []Role{
		RoleSuperAdmin,
		RoleAdmin,
		RoleAttendee,
		RolePremiumAttendee,
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
		ResourceBusiness,
		ResourceMember,
		ResourceUser,
		ResourceDashboard,
		ResourceAnalytics,
		ResourceNotification,
		ResourcePlatform,
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
	}
}