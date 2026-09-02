// internal/modules/auth/authorization/middleware.go

package authorization

import (
	"log"
	"net/http"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"

	"github.com/gofiber/fiber/v3"
)

// ============================================================
// CORE AUTHORIZATION MIDDLEWARE
// ============================================================

// AuthorizationMiddleware enforces permissions using Casbin
// This is the ONLY middleware you need for most routes
func AuthorizationMiddleware(checker authdomain.PermissionChecker) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user ID
		userID := c.Locals(authdomain.ContextKeyUserID)
		if userID == nil {
			return response.Unauthorized(c, "User not authenticated", fiber.Map{
				"reason": "user_id not found in context",
			})
		}

		userIDStr, ok := userID.(string)
		if !ok {
			return response.Unauthorized(c, "Invalid user ID", fiber.Map{
				"reason": "user_id is not a string",
			})
		}

		// Determine scope from request
		scope := getScopeFromRequest(c)

		// Determine resource and action
		resource := getResourceFromRequest(c)
		action := getActionFromRequest(c)

		// DEBUG
		log.Printf("🔍 AUTHZ: user=%s, scope=%s, resource=%s, action=%s",
			userIDStr, scope.String(), resource, action)

		// Store scope for downstream
		c.Locals("scope", scope)

		// Check permission
		allowed, err := checker.HasPermission(c.Context(), userIDStr, scope, resource, action)
		if err != nil {
			return response.InternalError(c, "Authorization error", fiber.Map{
				"error": err.Error(),
			})
		}

		if !allowed {
			roles, _ := checker.GetUserRoles(c.Context(), userIDStr, scope)
			return response.Forbidden(c, "Insufficient permissions", fiber.Map{
				"user":     userIDStr,
				"scope":    scope.String(),
				"resource": resource,
				"action":   action,
				"roles":    roles,
			})
		}

		return c.Next()
	}
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// getScopeFromRequest extracts the Scope from the request
func getScopeFromRequest(c fiber.Ctx) authdomain.Scope {
	path := c.Path()

	// Check context first
	if scope := c.Locals("scope"); scope != nil {
		if s, ok := scope.(authdomain.Scope); ok {
			return s
		}
	}

	// Check path params
	institutionID := c.Params("institutionId")
	if institutionID != "" {
		return authdomain.NewInstitutionTeamScope(institutionID)
	}

	userID := c.Params("userId")
	if userID != "" {
		return authdomain.NewPersonalTeamScope(userID)
	}

	teamID := c.Params("teamId")
	if teamID != "" {
		return authdomain.NewPersonalTeamScope(teamID)
	}

	id := c.Params("id")
	if id != "" {
		if strings.Contains(path, "/institutions/") {
			return authdomain.NewInstitutionTeamScope(id)
		}
		if strings.Contains(path, "/users/") || strings.Contains(path, "/teams/") {
			return authdomain.NewPersonalTeamScope(id)
		}
	}

	// Check /me endpoints
	if strings.Contains(path, "/me") || strings.Contains(path, "/my") {
		if uid := c.Locals(authdomain.ContextKeyUserID); uid != nil {
			if uidStr, ok := uid.(string); ok && uidStr != "" {
				return authdomain.NewPersonalTeamScope(uidStr)
			}
		}
	}

	// Check query params
	if institutionID = c.Query("institutionId"); institutionID != "" {
		return authdomain.NewInstitutionTeamScope(institutionID)
	}
	if userID = c.Query("userId"); userID != "" {
		return authdomain.NewPersonalTeamScope(userID)
	}
	if teamID = c.Query("teamId"); teamID != "" {
		return authdomain.NewPersonalTeamScope(teamID)
	}

	// Platform routes
	if strings.HasPrefix(path, "/api/v1/admin") ||
		strings.HasPrefix(path, "/api/v1/platform") ||
		strings.HasPrefix(path, "/api/v1/system") {
		return authdomain.NewPlatformScope()
	}

	// Default to user's personal team
	if uid := c.Locals(authdomain.ContextKeyUserID); uid != nil {
		if uidStr, ok := uid.(string); ok && uidStr != "" {
			return authdomain.NewPersonalTeamScope(uidStr)
		}
	}

	return authdomain.NewPlatformScope()
}

// getResourceFromRequest extracts the resource from the request path
func getResourceFromRequest(c fiber.Ctx) string {
	path := strings.TrimPrefix(c.Path(), "/api/v1/")
	segments := strings.Split(path, "/")

	for i, seg := range segments {
		switch seg {
		case "users", "user":
			if i+2 < len(segments) && (segments[i+2] == "events" || segments[i+2] == "event") {
				return authdomain.ResourceEvent.String()
			}
			return authdomain.ResourceUser.String()
		case "institutions", "institution":
			if i+2 < len(segments) && (segments[i+2] == "events" || segments[i+2] == "event") {
				return authdomain.ResourceEvent.String()
			}
			return authdomain.ResourceInstitution.String()
		case "teams", "team":
			if i+2 < len(segments) && (segments[i+2] == "events" || segments[i+2] == "event") {
				return authdomain.ResourceEvent.String()
			}
			return authdomain.ResourceTeam.String()
		case "me":
			if i+1 < len(segments) && (segments[i+1] == "events" || segments[i+1] == "event") {
				return authdomain.ResourceEvent.String()
			}
		}
	}

	// Check for direct matches
	for _, seg := range segments {
		switch seg {
		case "profile", "events", "event", "certificates", "certificate",
			"attendees", "attendee", "payments", "payment", "payouts", "payout",
			"dashboard", "analytics", "notifications", "notification",
			"media", "members", "member", "teams", "team", "team-types", "team-type":
			return seg
		}
	}

	return ""
}

// getActionFromRequest maps HTTP method to action
func getActionFromRequest(c fiber.Ctx) string {
	switch c.Method() {
	case http.MethodGet:
		return authdomain.ActionRead.String()
	case http.MethodPost:
		return authdomain.ActionCreate.String()
	case http.MethodPut, http.MethodPatch:
		return authdomain.ActionUpdate.String()
	case http.MethodDelete:
		return authdomain.ActionDelete.String()
	default:
		return authdomain.ActionRead.String()
	}
}