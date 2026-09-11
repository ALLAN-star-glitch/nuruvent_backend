// internal/modules/events/domain/context_keys.go

package domain

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// Alias the shared context key type
type ContextKey = types.ContextKey

// Re-export shared constants (so existing events code keeps working)
const (
	ContextKeyUserID      = types.ContextKeyUserID
	ContextKeyUserRole    = types.ContextKeyUserRole
	ContextKeyUserEmail   = types.ContextKeyUserEmail
	ContextKeyUserName    = types.ContextKeyUserName
	ContextKeyAccountID   = types.ContextKeyAccountID
	ContextKeyAccountType = types.ContextKeyAccountType
	ContextKeyTeamID      = types.ContextKeyTeamID
	ContextKeyTeamType    = types.ContextKeyTeamType
)

// ============================================================
// GO context.Context helpers (unchanged — separate from Fiber Locals)
// ============================================================

// WithUserID injects the user ID into the Go context.Context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// GetUserID extracts the user ID from the Go context.Context
func GetUserID(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return val
	}
	return ""
}

// WithTeamContext injects team ID and team type into the Go context.Context
func WithTeamContext(ctx context.Context, teamID, teamType string) context.Context {
	ctx = context.WithValue(ctx, ContextKeyTeamID, teamID)
	return context.WithValue(ctx, ContextKeyTeamType, teamType)
}

// GetTeamID extracts the team ID from the Go context.Context
func GetTeamID(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyTeamID).(string); ok {
		return val
	}
	return ""
}

// GetTeamType extracts the team type from the Go context.Context
func GetTeamType(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyTeamType).(string); ok {
		return val
	}
	return ""
}