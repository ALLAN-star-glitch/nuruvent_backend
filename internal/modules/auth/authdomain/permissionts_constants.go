// internal/modules/auth/authdomain/constants.go

package authdomain

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"

// Role represents a user role in the system
type Role string

type ContextKey = types.ContextKey

const (
	ContextKeyUserID       = types.ContextKeyUserID
	ContextKeyUserRole     = types.ContextKeyUserRole
	ContextKeyUserEmail    = types.ContextKeyUserEmail
	ContextKeyUserName     = types.ContextKeyUserName
	ContextKeyDomain       = types.ContextKeyDomain
	ContextKeyUserRoles    = types.ContextKeyUserRoles
	ContextKeyAccountID    = types.ContextKeyAccountID
	ContextKeyAccountType  = types.ContextKeyAccountType
	ContextKeyTeamID       = types.ContextKeyTeamID
	ContextKeyTeamType     = types.ContextKeyTeamType
	ContextKeyAccessToken  = types.ContextKeyAccessToken
	ContextKeyRefreshToken = types.ContextKeyRefreshToken
)

// ============================================================
// ROLE DEFINITIONS
// ============================================================

const (
	// Platform-level roles (Nuruvent staff)
	RoleSuperAdmin Role = "super_admin" // Full platform control
	RoleAdmin      Role = "admin"       // Platform management

	// Account-level roles (within accounts)
	RoleAccountAdmin Role = "account_admin" // Full account management
	RoleTrainer      Role = "trainer"       // Training focused

	// System
	RoleGuest Role = "guest" // Unregistered user
)

// ============================================================
// RESOURCE DEFINITIONS
// ============================================================

// Resource represents a resource being accessed
type Resource string

const (
	// Platform resources
	ResourcePlatform    Resource = "platform"
	ResourceUser        Resource = "user"
	ResourceInstitution Resource = "institution"
	ResourceTeam        Resource = "team"

	// Account resources
	ResourceAccount Resource = "account"
	ResourceBilling Resource = "billing"
	ResourceSetting Resource = "setting"

	// Team-level resources
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

// ============================================================
// ACTION DEFINITIONS
// ============================================================

// Action represents an operation that can be performed
type Action string

const (
	// Standard actions
	ActionCreate   Action = "create"
	ActionRead     Action = "read"
	ActionUpdate   Action = "update"
	ActionDelete   Action = "delete"
	ActionManage   Action = "manage"
	ActionIssue    Action = "issue"
	ActionRegister Action = "register"
	ActionExport   Action = "export"
	ActionRefund   Action = "refund"
	ActionDownload Action = "download"
	ActionInvite   Action = "invite"
	ActionViewCreator Action = "view_creator"

	// Account-specific actions
	ActionMemberAdd     Action = "member_add"
	ActionMemberRemove  Action = "member_remove"
	ActionBillingRead   Action = "billing_read"
	ActionBillingUpdate Action = "billing_update"

	// OWN vs ALL Actions
	ActionReadAll    Action = "read_all"
	ActionReadOwn    Action = "read_own"
	ActionUpdateAll  Action = "update_all"
	ActionUpdateOwn  Action = "update_own"
	ActionDeleteAll  Action = "delete_all"
	ActionDeleteOwn  Action = "delete_own"
	ActionPublishAll Action = "publish_all"
	ActionPublishOwn Action = "publish_own"
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

func (a Action) IsOwnAction() bool {
	switch a {
	case ActionReadOwn, ActionUpdateOwn, ActionDeleteOwn, ActionPublishOwn:
		return true
	default:
		return false
	}
}

func (a Action) IsAllAction() bool {
	switch a {
	case ActionReadAll, ActionUpdateAll, ActionDeleteAll, ActionPublishAll:
		return true
	default:
		return false
	}
}

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
// ROLE VALIDATION HELPERS
// ============================================================

func IsValidRole(role string) bool {
	validRoles := map[string]bool{
		RoleSuperAdmin.String():   true,
		RoleAdmin.String():        true,
		RoleAccountAdmin.String(): true,
		RoleTrainer.String():      true,
		RoleGuest.String():        true,
	}
	return validRoles[role]
}

func IsPlatformRole(role string) bool {
	platformRoles := map[string]bool{
		RoleSuperAdmin.String(): true,
		RoleAdmin.String():      true,
		RoleGuest.String():      true,
	}
	return platformRoles[role]
}

func IsAccountRole(role string) bool {
	accountRoles := map[string]bool{
		RoleAccountAdmin.String(): true,
		RoleTrainer.String():      true,
	}
	return accountRoles[role]
}

func IsValidAccountRole(role string) bool {
	return IsAccountRole(role)
}

// ============================================================
// GET ALL HELPERS
// ============================================================

func GetAllRoles() []Role {
	return []Role{
		RoleSuperAdmin,
		RoleAdmin,
		RoleAccountAdmin,
		RoleTrainer,
		RoleGuest,
	}
}

func GetAllPlatformRoles() []Role {
	return []Role{
		RoleSuperAdmin,
		RoleAdmin,
		RoleGuest,
	}
}

func GetAllAccountRoles() []Role {
	return []Role{
		RoleAccountAdmin,
		RoleTrainer,
	}
}

func GetAllResources() []Resource {
	return []Resource{
		ResourcePlatform,
		ResourceUser,
		ResourceInstitution,
		ResourceTeam,
		ResourceAccount,
		ResourceBilling,
		ResourceSetting,
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

func GetAllAccountResources() []Resource {
	return []Resource{
		ResourceAccount,
		ResourceBilling,
		ResourceSetting,
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
		ActionMemberAdd,
		ActionMemberRemove,
		ActionBillingRead,
		ActionBillingUpdate,
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
// DEFAULT PERMISSION MATRIX
// ============================================================

func DefaultPlatformPermissions() map[Role][]string {
	return map[Role][]string{
		RoleSuperAdmin: {
			"*:*",
		},
		RoleAdmin: {
			"platform:read",
			"platform:update",
			"user:read",
			"user:update",
			"user:delete",
			"institution:read",
			"institution:update",
			"institution:delete",
			"account:read",
			"account:update",
			"account:delete",
			"event:read_all",
			"event:update_all",
			"event:delete_all",
			"team:read",
			"team:update",
			"team:delete",
			"analytics:read",
			"billing:read",
			"certificate:read",
			"member:read",
			"member:delete",
			"profile:read_all",
			"profile:update_all",
		},
	}
}

func DefaultAccountPermissions() map[Role][]string {
	return map[Role][]string{
		RoleAccountAdmin: {
			// Account management
			"account:read",
			"account:update",
			"account:delete",
			"account:member_add",
			"account:member_remove",
			"billing:read",
			"billing:update",
			"setting:read",
			"setting:update",

			// Team management
			"team:create",
			"team:read",
			"team:update",
			"team:delete",

			// Event management
			"event:create",
			"event:read_all",
			"event:update_all",
			"event:delete_all",
			"event:publish_all",
			"event:manage",

			// Member management
			"member:create",
			"member:read",
			"member:update",
			"member:delete",
			"member:invite",
	

			// Profile
			"profile:read_all",
			"profile:update_all",

			// Certificate
			"certificate:create",
			"certificate:read",
			"certificate:update",
			"certificate:delete",
			"certificate:issue",

			// Attendee
			"attendee:read",
			"attendee:update",
			"attendee:export",

			// Dashboard
			"dashboard:read",
			"analytics:read",
		},
		RoleTrainer: {
			// Events
			"event:create",
			"event:read_all",
			"event:update_all",
			"event:delete_all",
			"event:publish_all",

			// Attendee
			"attendee:create",
			"attendee:read",
			"attendee:update",
			"attendee:delete",
			"attendance:track",
			"attendance:export",

			// Certificate
			"certificate:create",
			"certificate:read",
			"certificate:issue",
			"certificate:revoke",
			"certificate:download",

			// Course materials
			"material:create",
			"material:read",
			"material:update",
			"material:delete",
			"material:upload",

			// Analytics
			"analytics:read_training",
			"analytics:read_attendance",
			"analytics:read_certificates",
			"analytics:export_training",

			// Profile
			"profile:read",
			"profile:update",

			// Team
			"team:read",

			// Account
			"account:read",
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
		RoleTrainer:      70,
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