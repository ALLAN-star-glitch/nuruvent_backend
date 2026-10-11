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
	RoleLearner      Role = "learner"       // Consumes training content; no management

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
	ResourceRegistration Resource = "registration"
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
	// ---- Standard actions ----
	ActionCreate      Action = "create"
	ActionRead        Action = "read"
	ActionUpdate      Action = "update"
	ActionDelete      Action = "delete"
	ActionManage      Action = "manage"
	ActionIssue       Action = "issue"
	ActionRegister    Action = "register"
	ActionExport      Action = "export"
	ActionRefund      Action = "refund"
	ActionDownload    Action = "download"
	ActionInvite      Action = "invite"
	ActionViewCreator Action = "view_creator"

	// ---- Account-specific actions ----
	ActionMemberAdd     Action = "member_add"
	ActionMemberRemove  Action = "member_remove"
	ActionBillingRead   Action = "billing_read"
	ActionBillingUpdate Action = "billing_update"

	// ---- Membership / self-service actions ----
	ActionLeave Action = "leave" // member self-removal from an account/team

	// ---- OWN vs ALL actions ----
	// These encode data-scope for a resource: whether the grant applies to
	// the whole account or only to rows the caller created / is associated
	// with. The service layer applies the actual query filter; the policy
	// grant expresses which scope the caller may operate within.
	ActionReadAll    Action = "read_all"
	ActionReadOwn    Action = "read_own"
	ActionUpdateAll  Action = "update_all"
	ActionUpdateOwn  Action = "update_own"
	ActionDeleteAll  Action = "delete_all"
	ActionDeleteOwn  Action = "delete_own"
	ActionPublishAll Action = "publish_all"
	ActionPublishOwn Action = "publish_own"
	ActionExportOwn  Action = "export_own"
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

// IsOwnAction reports whether the action is scoped to the caller's own
// rows rather than the whole account.
func (a Action) IsOwnAction() bool {
	switch a {
	case ActionReadOwn, ActionUpdateOwn, ActionDeleteOwn, ActionPublishOwn, ActionExportOwn:
		return true
	default:
		return false
	}
}

// IsAllAction reports whether the action is scoped to the whole account.
func (a Action) IsAllAction() bool {
	switch a {
	case ActionReadAll, ActionUpdateAll, ActionDeleteAll, ActionPublishAll:
		return true
	default:
		return false
	}
}

// GetOwnAction maps a base action to its own-scoped counterpart.
// Returns the input unchanged when no own-scoped counterpart exists.
func (a Action) GetOwnAction() Action {
	switch a {
	case ActionRead:
		return ActionReadOwn
	case ActionUpdate:
		return ActionUpdateOwn
	case ActionDelete:
		return ActionDeleteOwn
	case ActionExport:
		return ActionExportOwn
	case ActionPublishAll:
		return ActionPublishOwn
	default:
		return a
	}
}

// GetAllAction maps a base or own-scoped action to its account-wide
// counterpart. Returns the input unchanged when no all-scoped
// counterpart exists.
func (a Action) GetAllAction() Action {
	switch a {
	case ActionRead:
		return ActionReadAll
	case ActionUpdate:
		return ActionUpdateAll
	case ActionDelete:
		return ActionDeleteAll
	case ActionExportOwn:
		return ActionExport
	case ActionPublishOwn:
		return ActionPublishAll
	default:
		return a
	}
}

// ============================================================
// ROLE VALIDATION HELPERS
// ============================================================

// IsValidRole reports whether the string is a known role.
func IsValidRole(role string) bool {
	validRoles := map[string]bool{
		RoleSuperAdmin.String():   true,
		RoleAdmin.String():        true,
		RoleAccountAdmin.String(): true,
		RoleTrainer.String():      true,
		RoleLearner.String():      true,
		RoleGuest.String():        true,
	}
	return validRoles[role]
}

// IsPlatformRole reports whether the role belongs to the platform
// domain (Nuruvent staff and unauthenticated guests).
func IsPlatformRole(role string) bool {
	platformRoles := map[string]bool{
		RoleSuperAdmin.String(): true,
		RoleAdmin.String():      true,
		RoleGuest.String():      true,
	}
	return platformRoles[role]
}

// IsAccountRole reports whether the role belongs to the account domain.
func IsAccountRole(role string) bool {
	accountRoles := map[string]bool{
		RoleAccountAdmin.String(): true,
		RoleTrainer.String():      true,
		RoleLearner.String():      true,
	}
	return accountRoles[role]
}

// IsValidAccountRole is an alias for IsAccountRole. Kept for callers
// that read better as "is this a valid account role" in context.
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
		RoleLearner,
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
		RoleLearner,
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
		ResourceRegistration,
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
		ActionLeave,
		ActionReadAll,
		ActionReadOwn,
		ActionUpdateAll,
		ActionUpdateOwn,
		ActionDeleteAll,
		ActionDeleteOwn,
		ActionPublishAll,
		ActionPublishOwn,
		ActionExportOwn,
	}
}

// ============================================================
// DEFAULT PERMISSION MATRIX
// ============================================================
//
// These two functions are descriptive, not authoritative. The
// authoritative policy set lives in authorization.GetPlatformPolicies()
// and authorization.GetAccountPolicies().
//
// They exist so tooling and documentation can present a flat
// "what does role X get" view without reading the policy builder. They
// MUST stay in sync with the seeded policy set. Any change to
// authorization.GetPlatformPolicies() or
// authorization.GetAccountPolicies() must be mirrored here, or the
// two views diverge and this file becomes a source of confusion
// rather than clarity.
//
// ValidatePolicies() enforces the policy invariants at seed time;
// keeping these maps aligned is a review responsibility.

func DefaultPlatformPermissions() map[Role][]string {
	return map[Role][]string{
		RoleSuperAdmin: {
			"*:*",
		},
		RoleAdmin: {
			"platform:manage",
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
			"event:manage",
			"analytics:read",
			"payment:read",
			"certificate:read",
			"member:read",
			"member:delete",
			"media:read",
			"media:delete",
			"team:read",
			"team:update",
			"team:delete",
		},
	}
}

func DefaultAccountPermissions() map[Role][]string {
	return map[Role][]string{
		// ============================================================
		// ACCOUNT ADMIN
		// ============================================================
		// Full account control. No scoping — admins operate on every row
		// in the account.
		RoleAccountAdmin: {
			// Account
			"account:read",
			"account:update",
			"account:delete",


			"registration:read",
			"registration:update",
			"registration:delete",
			"registration:export",

			// Members
			"member:create",
			"member:read",
			"member:update",
			"member:delete",
			"member:invite",
			"member:leave",

			// Teams
			"team:create",
			"team:read",
			"team:update",
			"team:delete",
			"team:manage",

			// Events
			"event:create",
			"event:read_all",
			"event:update_all",
			"event:delete_all",
			"event:publish_all",
			"event:manage",

			// Certificates
			"certificate:create",
			"certificate:read",
			"certificate:update",
			"certificate:issue",
			"certificate:delete",

			// Attendees
			"attendee:read",
			"attendee:update",
			"attendee:export",

			// Payments
			"payment:read",
			"payment:create",
			"payment:refund",

			// Billing
			"billing:read",
			"billing:update",

			// Settings
			"setting:read",
			"setting:update",

			// Dashboard, profile, analytics
			"dashboard:read",
			"profile:read",
			"profile:update",
			"analytics:read",
		},

		// ============================================================
		// TRAINER
		// ============================================================
		// Operational staff. Account-wide read of the event roster, but
		// manages only their own events. Payments and attendees are
		// own-scoped. Member roster is visible (team-aware filtering is
		// applied at the service layer for team endpoints).
		RoleTrainer: {
			// Account (read only)
			"account:read",

			// Members (view roster, invite, leave)
			"member:read",
			"member:invite",
			"member:leave",


			"registration:read_own",
			"registration:update_own",
			"registration:export_own",

			// Teams (read only)
			"team:read",

			// Events: account-wide read, own-only write
			"event:create",
			"event:read_all",
			"event:update_own",
			"event:delete_own",
			"event:publish_own",

			// Attendees: own events only
			"attendee:create",
			"attendee:read_own",
			"attendee:update_own",
			"attendee:export_own",

			// Certificates
			"certificate:create",
			"certificate:read",
			"certificate:issue",

	

			// Dashboard, profile, analytics
			"dashboard:read",
			"profile:read",
			"profile:update",
			"analytics:read",
		},

		// ============================================================
		// LEARNER
		// ============================================================
		// Consumes training content. No management anywhere in the
		// account. Visibility is scoped by the service layer:
		//
		//   - Events: only events on teams the learner belongs to.
		//   - Attendees: only the learner's own records.
		//   - Certificates: only the learner's own certificates.
		//   - Members: no account-wide roster. Team rosters are read via
		//     the team members endpoint, which is scoped to teams the
		//     learner belongs to.
		//
		// The learner does NOT hold `member:read` or `account:read`. The
		// UserMenu role badge reads from the self-service
		// GET /accounts/:id/members/me endpoint, which requires no
		// member-list grant.
		RoleLearner: {
			// Teams (read only)
			"team:read",

			// Members (team-scoped read, self-service leave)
			"member:read_own",
			"member:leave",

			"registration:read_own",


			// Events: visible-scope read + register
			"event:read_own",
			"event:register",

			// Attendee: own records only
			"attendee:read_own",
			"attendee:update_own",

			// Certificates: own records only
			"certificate:read",
			"certificate:download",

			// Profile
			"profile:read",
			"profile:update",

			// Dashboard
			"dashboard:read",
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
		RoleLearner:      50,
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