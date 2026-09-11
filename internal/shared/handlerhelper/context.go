// internal/shared/handlerhelper/context.go

package handlerhelper

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// FIBER LOCALS READERS
// ============================================================
//
// These read from Fiber's c.Locals() — the request-scoped key-value store
// that auth/authz middleware populate. They are the Fiber-side counterpart
// of the Go-context helpers in shared/types.
//
// Every handler that needs request context (user ID, team ID, account ID)
// reads it through these helpers rather than accessing c.Locals directly.
// That keeps the key strings in one place and makes refactoring safe.

// ---- User ----

// GetUserID extracts the authenticated user ID from c.Locals.
// Returns an error if the user is not authenticated.
func GetUserID(c fiber.Ctx) (string, error) {
	userID := c.Locals(types.ContextKeyUserID)
	if userID == nil {
		return "", errors.New("user not authenticated")
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		return "", errors.New("invalid user ID")
	}
	return userIDStr, nil
}

// GetUserIDOptional extracts the user ID if present, otherwise returns "".
// Use this on routes where authentication is optional.
func GetUserIDOptional(c fiber.Ctx) string {
	if user := c.Locals(types.ContextKeyUserID); user != nil {
		if id, ok := user.(string); ok {
			return id
		}
	}
	return ""
}

// GetUserRole extracts the user's role from c.Locals.
func GetUserRole(c fiber.Ctx) string {
	if v := c.Locals(types.ContextKeyUserRole); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetUserEmail extracts the user's email from c.Locals.
func GetUserEmail(c fiber.Ctx) string {
	if v := c.Locals(types.ContextKeyUserEmail); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetUserName extracts the user's display name from c.Locals.
func GetUserName(c fiber.Ctx) string {
	if v := c.Locals(types.ContextKeyUserName); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ---- Account ----

// GetAccountID extracts the account ID from c.Locals.
func GetAccountID(c fiber.Ctx) string {
	if v := c.Locals(types.ContextKeyAccountID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetAccountType extracts the account type from c.Locals.
func GetAccountType(c fiber.Ctx) string {
	if v := c.Locals(types.ContextKeyAccountType); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ---- Team ----

// GetTeamID extracts the team ID from c.Locals.
func GetTeamID(c fiber.Ctx) string {
	if v := c.Locals(types.ContextKeyTeamID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetTeamType extracts the team type ("personal" | "institution") from c.Locals.
func GetTeamType(c fiber.Ctx) string {
	if v := c.Locals(types.ContextKeyTeamType); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ---- Domain (permission scope resolved by authz middleware) ----

// GetDomain extracts the resolved permission domain from c.Locals.
func GetDomain(c fiber.Ctx) string {
	if v := c.Locals(types.ContextKeyDomain); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ============================================================
// CONTEXT BRIDGING
// ============================================================

// EnrichUserContext syncs Fiber's c.Locals values into a standard Go
// context.Context, so downstream service code can read them via
// shared/types.GetUserID(ctx), types.GetTeamID(ctx), etc.
//
// Populates (only when present):
//   - user_id
//   - team_id + team_type
//   - account_id + account_type
//   - domain
//
// Call this at the top of any handler that will pass its context to a
// service that performs permission checks or scopes queries by user/team.
func EnrichUserContext(c fiber.Ctx) context.Context {
	ctx := c.Context()

	if userID := GetUserIDOptional(c); userID != "" {
		ctx = types.WithUserID(ctx, userID)
	}

	if teamID := GetTeamID(c); teamID != "" {
		ctx = types.WithTeamContext(ctx, teamID, GetTeamType(c))
	}

	if accountID := GetAccountID(c); accountID != "" {
		ctx = types.WithAccountContext(ctx, accountID, GetAccountType(c))
	}

	if domain := GetDomain(c); domain != "" {
		ctx = types.WithDomain(ctx, domain)
	}

	return ctx
}