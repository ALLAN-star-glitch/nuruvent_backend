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

		// ✅ SPECIAL CASE: /users/me/profile - always allow
		// The service handles permission checks internally (viewerID == userID)
		if isOwnProfileRequest(c) {
			log.Printf("✅ AUTHZ BYPASS: /users/me/profile - allowing access")
			c.Locals("scope", scope)
			return c.Next()
		}

		// Store scope for downstream
		c.Locals("scope", scope)

		// Check permission with fallback chain
		allowed, err := checkPermissionWithFallback(c, checker, userIDStr, scope, resource, action)
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

// isOwnProfileRequest checks if this is a /users/me/profile request
func isOwnProfileRequest(c fiber.Ctx) bool {
	path := c.Path()
	method := c.Method()
	return method == http.MethodGet && strings.Contains(path, "/users/me/profile")
}

// checkPermissionWithFallback checks permissions with a fallback chain
func checkPermissionWithFallback(
	c fiber.Ctx,
	checker authdomain.PermissionChecker,
	userID string,
	scope authdomain.Scope,
	resource string,
	action string,
) (bool, error) {
	ctx := c.Context()

	// For read actions: read -> read_all -> read_own
	if action == authdomain.ActionRead.String() {
		// Try exact read permission
		allowed, err := checker.HasPermission(ctx, userID, scope, resource, authdomain.ActionRead.String())
		if err == nil && allowed {
			return true, nil
		}

		// Try read_all
		allowed, err = checker.HasPermission(ctx, userID, scope, resource, authdomain.ActionReadAll.String())
		if err == nil && allowed {
			return true, nil
		}

		// Try read_own
		allowed, err = checker.HasPermission(ctx, userID, scope, resource, authdomain.ActionReadOwn.String())
		if err == nil && allowed {
			return true, nil
		}

		if err != nil {
			return false, err
		}
		return false, nil
	}

	// For update actions: update -> update_all -> update_own
	if action == authdomain.ActionUpdate.String() {
		// Try exact update permission
		allowed, err := checker.HasPermission(ctx, userID, scope, resource, authdomain.ActionUpdate.String())
		if err == nil && allowed {
			return true, nil
		}

		// Try update_all
		allowed, err = checker.HasPermission(ctx, userID, scope, resource, authdomain.ActionUpdateAll.String())
		if err == nil && allowed {
			return true, nil
		}

		// Try update_own
		allowed, err = checker.HasPermission(ctx, userID, scope, resource, authdomain.ActionUpdateOwn.String())
		if err == nil && allowed {
			return true, nil
		}

		if err != nil {
			return false, err
		}
		return false, nil
	}

	// For delete actions: delete -> delete_all -> delete_own
	if action == authdomain.ActionDelete.String() {
		// Try exact delete permission
		allowed, err := checker.HasPermission(ctx, userID, scope, resource, authdomain.ActionDelete.String())
		if err == nil && allowed {
			return true, nil
		}

		// Try delete_all
		allowed, err = checker.HasPermission(ctx, userID, scope, resource, authdomain.ActionDeleteAll.String())
		if err == nil && allowed {
			return true, nil
		}

		// Try delete_own
		allowed, err = checker.HasPermission(ctx, userID, scope, resource, authdomain.ActionDeleteOwn.String())
		if err == nil && allowed {
			return true, nil
		}

		if err != nil {
			return false, err
		}
		return false, nil
	}

	// For publish actions: publish_all -> publish_own (no exact publish)
	if action == authdomain.ActionPublishAll.String() || action == authdomain.ActionPublishOwn.String() {
		// Try publish_all
		allowed, err := checker.HasPermission(ctx, userID, scope, resource, authdomain.ActionPublishAll.String())
		if err == nil && allowed {
			return true, nil
		}

		// Try publish_own
		allowed, err = checker.HasPermission(ctx, userID, scope, resource, authdomain.ActionPublishOwn.String())
		if err == nil && allowed {
			return true, nil
		}

		if err != nil {
			return false, err
		}
		return false, nil
	}

	// For create actions: just check create
	if action == authdomain.ActionCreate.String() {
		return checker.HasPermission(ctx, userID, scope, resource, authdomain.ActionCreate.String())
	}

	// For manage actions: just check manage
	if action == authdomain.ActionManage.String() {
		return checker.HasPermission(ctx, userID, scope, resource, authdomain.ActionManage.String())
	}

	// For all other actions, check exact match
	return checker.HasPermission(ctx, userID, scope, resource, action)
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

	// Check for profile endpoints
	if strings.Contains(path, "/profile") {
		// For /users/me/profile - use personal scope
		if strings.Contains(path, "/users/me/profile") {
			if uid := c.Locals(authdomain.ContextKeyUserID); uid != nil {
				if uidStr, ok := uid.(string); ok && uidStr != "" {
					return authdomain.NewPersonalTeamScope(uidStr)
				}
			}
		}
		// For /profile/organizer?scope=xxx - parse from query
		if strings.Contains(path, "/profile/organizer") {
			scopeParam := c.Query("scope")
			if scopeParam != "" {
				parts := strings.SplitN(scopeParam, ":", 2)
				if len(parts) == 2 {
					switch parts[0] {
					case "personal":
						return authdomain.NewPersonalTeamScope(parts[1])
					case "institution":
						return authdomain.NewInstitutionTeamScope(parts[1])
					}
				}
			}
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
			// Check if this is a profile endpoint
			if i+1 < len(segments) && (segments[i+1] == "profile" || strings.Contains(segments[i+1], "profile")) {
				return "profile"
			}
			if i+2 < len(segments) && (segments[i+2] == "events" || segments[i+2] == "event") {
				return authdomain.ResourceEvent.String()
			}
			if i+2 < len(segments) && segments[i+2] == "profiles" {
				return "profile"
			}
			return authdomain.ResourceUser.String()
		case "profile":
			return "profile"
		case "institutions", "institution":
			if i+2 < len(segments) && (segments[i+2] == "events" || segments[i+2] == "event") {
				return authdomain.ResourceEvent.String()
			}
			if i+2 < len(segments) && segments[i+2] == "profile" {
				return "profile"
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
			if i+1 < len(segments) && segments[i+1] == "profile" {
				return "profile"
			}
		}
	}

	// Check for direct matches
	for _, seg := range segments {
		switch seg {
		case "profile":
			return "profile"
		case "events", "event":
			return authdomain.ResourceEvent.String()
		case "certificates", "certificate":
			return authdomain.ResourceCertificate.String()
		case "attendees", "attendee":
			return authdomain.ResourceAttendee.String()
		case "payments", "payment":
			return authdomain.ResourcePayment.String()
		case "payouts", "payout":
			return authdomain.ResourcePayout.String()
		case "members", "member":
			return authdomain.ResourceMember.String()
		case "dashboard":
			return "dashboard"
		case "analytics":
			return "analytics"
		case "notifications", "notification":
			return "notification"
		case "media":
			return "media"
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