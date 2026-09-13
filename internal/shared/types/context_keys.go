// internal/shared/types/context_keys.go

package types

import "context"

// ============================================================
// CONTEXT KEY TYPE
// ============================================================

// ContextKey is the shared key type for values stored in Fiber's c.Locals()
// and in the standard library context.Context. Using a named type (rather
// than plain string) prevents accidental assignment from unrelated strings
// and catches type mismatches at compile time.

type ContextKey string

// ============================================================
// CONTEXT KEY CONSTANTS
// ============================================================

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

	// Tokens (used in auth middleware that reads/writes via Locals)
	ContextKeyAccessToken  ContextKey = "access_token"
	ContextKeyRefreshToken ContextKey = "refresh_token"
)

// ============================================================
// GO context.Context HELPERS
// ============================================================
//
// These helpers read/write values from the standard library context.Context.
// They mirror what Fiber's c.Locals() does for the request-scoped locals map
// — but for the Go context that flows into the service/domain layer.
//
// The middleware bridges the two: it reads from c.Locals (Fiber) and writes
// into the Go context (via WithUserID / WithTeamContext), so downstream
// service code only depends on the Go context.

// ---- User ----

// WithUserID injects the authenticated user ID into the Go context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// GetUserID extracts the user ID from the Go context.
// Returns "" if not set or not a string.
func GetUserID(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return val
	}
	return ""
}

// ---- Team ----

// WithTeamContext injects team ID and team type into the Go context.
func WithTeamContext(ctx context.Context, teamID, teamType string) context.Context {
	ctx = context.WithValue(ctx, ContextKeyTeamID, teamID)
	return context.WithValue(ctx, ContextKeyTeamType, teamType)
}


// GetTeamID extracts the team ID from the Go context.
func GetTeamID(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyTeamID).(string); ok {
		return val
	}
	return ""
}

// GetTeamType extracts the team type ("personal" | "institution") from the Go context.
func GetTeamType(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyTeamType).(string); ok {
		return val
	}
	return ""
}

// ---- Account ----

// WithAccountContext injects account ID and account type into the Go context.
func WithAccountContext(ctx context.Context, accountID, accountType string) context.Context {
	ctx = context.WithValue(ctx, ContextKeyAccountID, accountID)
	return context.WithValue(ctx, ContextKeyAccountType, accountType)
}

// GetAccountID extracts the account ID from the Go context.
func GetAccountID(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyAccountID).(string); ok {
		return val
	}
	return ""
}

// GetAccountType extracts the account type from the Go context.
func GetAccountType(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyAccountType).(string); ok {
		return val
	}
	return ""
}

// ---- Domain (permission scope resolved by authorization middleware) ----

// WithDomain injects the resolved permission domain into the Go context.
func WithDomain(ctx context.Context, domain string) context.Context {
	return context.WithValue(ctx, ContextKeyDomain, domain)
}

// GetDomain extracts the resolved permission domain from the Go context.
func GetDomain(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyDomain).(string); ok {
		return val
	}
	return ""
}