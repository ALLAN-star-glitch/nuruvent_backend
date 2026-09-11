// internal/shared/types/context_keys.go

package types

// ============================================================
// SHARED CONTEXT KEYS
// ============================================================
//
// These keys are used by auth middleware to inject claims into Fiber's
// c.Locals(), and by all downstream modules (events, teams, media, etc.)
// to read them back out.
//
// IMPORTANT: Never redefine these in module-local packages. Always import
// from this shared package so every module reads/writes the SAME key type
// and string value. Mismatched types (e.g. two separate `type ContextKey string`
// declarations) will cause Fiber's c.Locals() lookups to silently miss.

// ContextKey is the shared key type for Fiber's c.Locals() map.
type ContextKey string


const (
	// User identity
	ContextKeyUserID    ContextKey = "user_id"
	ContextKeyUserRole  ContextKey = "user_role"
	ContextKeyUserRoles ContextKey = "user_roles" // multi-role variant
	ContextKeyUserEmail ContextKey = "user_email"
	ContextKeyUserName  ContextKey = "user_name"

	// Authorization scope
	ContextKeyDomain ContextKey = "domain"

	// Account scope
	ContextKeyAccountID   ContextKey = "account_id"
	ContextKeyAccountType ContextKey = "account_type"

	// Team scope
	ContextKeyTeamID   ContextKey = "team_id"
	ContextKeyTeamType ContextKey = "team_type"

	// Tokens
	ContextKeyAccessToken  ContextKey = "access_token"
	ContextKeyRefreshToken ContextKey = "refresh_token"
)