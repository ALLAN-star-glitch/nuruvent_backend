package eventhandler

import (
	"context"
	"errors"
	"strconv"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/gofiber/fiber/v3"
)

// ============================================================
// QUERY PARSING HELPERS
// ============================================================

// getQueryInt parses an integer query parameter with a default fallback
func getQueryInt(c fiber.Ctx, key string, defaultValue int) int {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}

// getQueryString parses a string query parameter with a default fallback
func getQueryString(c fiber.Ctx, key string, defaultValue string) string {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// ============================================================
// USER CONTEXT HELPERS
// ============================================================

// getUserID extracts the authenticated user ID from c.Locals (returns error if missing)
func getUserID(c fiber.Ctx) (string, error) {
	userID := c.Locals(domain.ContextKeyUserID)
	if userID == nil {
		return "", errors.New("user not authenticated")
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		return "", errors.New("invalid user ID")
	}
	return userIDStr, nil
}

// getUserIDOptional extracts the user ID if present, otherwise returns empty string
func getUserIDOptional(c fiber.Ctx) string {
	if user := c.Locals(domain.ContextKeyUserID); user != nil {
		if id, ok := user.(string); ok {
			return id
		}
	}
	return ""
}

// ============================================================
// ACCOUNT & TEAM CONTEXT HELPERS
// ============================================================

// getAccountID extracts the Account ID from c.Locals if set by auth middleware
func getAccountID(c fiber.Ctx) string {
	if acc := c.Locals(domain.ContextKeyAccountID); acc != nil {
		if id, ok := acc.(string); ok {
			return id
		}
	}
	return ""
}

// getTeamID extracts the Team ID from c.Locals if set by auth middleware
func getTeamID(c fiber.Ctx) string {
	if team := c.Locals(domain.ContextKeyTeamID); team != nil {
		if id, ok := team.(string); ok {
			return id
		}
	}
	return ""
}



// getTeamType extracts the Team Type from c.Locals ("personal" or "institution")
func getTeamType(c fiber.Ctx) string {
	if tt := c.Locals(domain.ContextKeyTeamType); tt != nil {
		if teamType, ok := tt.(string); ok {
			return teamType
		}
	}
	return ""
}

// ============================================================
// CONTEXT BRIDGING HELPERS
// ============================================================

// enrichUserContext syncs Fiber's c.Locals values into a standard Go context.Context
// for passing down to the domain/service layer cleanly.
func enrichUserContext(c fiber.Ctx) context.Context {
	ctx := c.Context()

	if userID := getUserIDOptional(c); userID != "" {
		ctx = domain.WithUserID(ctx, userID)
	}

	teamID := getTeamID(c)
	teamType := getTeamType(c)
	if teamID != "" {
		ctx = domain.WithTeamContext(ctx, teamID, teamType)
	}

	return ctx
}