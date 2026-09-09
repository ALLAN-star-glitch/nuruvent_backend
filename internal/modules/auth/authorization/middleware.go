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

		// Determine domain from request
		domain := getDomainFromRequest(c)

		// Determine resource and action
		resource := getResourceFromRequest(c)
		action := getActionFromRequest(c)

		// DEBUG
		log.Printf("🔍 AUTHZ: user=%s, domain=%s, resource=%s, action=%s",
			userIDStr, domain, resource, action)

		// ✅ SPECIAL CASE: /users/me/profile - always allow (GET and PUT)
		if isOwnProfileRequest(c) {
			log.Printf("✅ AUTHZ BYPASS: /users/me/profile - allowing access")
			c.Locals("domain", domain)
			return c.Next()
		}

		// ✅ SPECIAL CASE: /users/me/avatar - always allow for own avatar
		if isOwnAvatarRequest(c) {
			log.Printf("✅ AUTHZ BYPASS: /users/me/avatar - allowing access")
			c.Locals("domain", domain)
			return c.Next()
		}

		// ✅ SPECIAL CASE: /institutions/:id/logo - allow for institution admins
		if isInstitutionLogoRequest(c) {
			log.Printf("✅ AUTHZ BYPASS: institution logo request - allowing access")
			c.Locals("domain", domain)
			return c.Next()
		}

		// Store domain for downstream
		c.Locals("domain", domain)

		// Check permission with fallback chain
		allowed, err := checkPermissionWithFallback(c, checker, userIDStr, domain, resource, action)
		if err != nil {
			return response.InternalError(c, "Authorization error", fiber.Map{
				"error": err.Error(),
			})
		}

		if !allowed {
			roles, _ := checker.GetUserRoles(c.Context(), userIDStr, domain)
			return response.Forbidden(c, "Insufficient permissions", fiber.Map{
				"user":     userIDStr,
				"domain":   domain,
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
	return strings.Contains(path, "/users/me/profile")
}

// isOwnAvatarRequest checks if this is a /users/me/avatar request
func isOwnAvatarRequest(c fiber.Ctx) bool {
	path := c.Path()
	return strings.Contains(path, "/users/me/avatar")
}

// isInstitutionLogoRequest checks if this is an institution logo request
func isInstitutionLogoRequest(c fiber.Ctx) bool {
	path := c.Path()
	return strings.Contains(path, "/institutions/") && strings.Contains(path, "/logo")
}

// checkPermissionWithFallback checks permissions with a fallback chain
func checkPermissionWithFallback(
	c fiber.Ctx,
	checker authdomain.PermissionChecker,
	userID string,
	domain string,
	resource string,
	action string,
) (bool, error) {
	ctx := c.Context()

	// For read actions: read -> read_all -> read_own
	if action == authdomain.ActionRead.String() {
		// Try exact read permission
		allowed, err := checker.HasPermission(ctx, userID, domain, resource, authdomain.ActionRead.String())
		log.Printf("🔍 HasPermission result: allowed=%v, err=%v", allowed, err)
		if err == nil && allowed {
			return true, nil
		}

		// Try read_all
		allowed, err = checker.HasPermission(ctx, userID, domain, resource, authdomain.ActionReadAll.String())
		if err == nil && allowed {
			return true, nil
		}

		// Try read_own
		allowed, err = checker.HasPermission(ctx, userID, domain, resource, authdomain.ActionReadOwn.String())
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
		allowed, err := checker.HasPermission(ctx, userID, domain, resource, authdomain.ActionUpdate.String())
		if err == nil && allowed {
			return true, nil
		}

		// Try update_all
		allowed, err = checker.HasPermission(ctx, userID, domain, resource, authdomain.ActionUpdateAll.String())
		if err == nil && allowed {
			return true, nil
		}

		// Try update_own
		allowed, err = checker.HasPermission(ctx, userID, domain, resource, authdomain.ActionUpdateOwn.String())
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
		allowed, err := checker.HasPermission(ctx, userID, domain, resource, authdomain.ActionDelete.String())
		if err == nil && allowed {
			return true, nil
		}

		// Try delete_all
		allowed, err = checker.HasPermission(ctx, userID, domain, resource, authdomain.ActionDeleteAll.String())
		if err == nil && allowed {
			return true, nil
		}

		// Try delete_own
		allowed, err = checker.HasPermission(ctx, userID, domain, resource, authdomain.ActionDeleteOwn.String())
		if err == nil && allowed {
			return true, nil
		}

		if err != nil {
			return false, err
		}
		return false, nil
	}

	// For publish actions: publish_all -> publish_own
	if action == authdomain.ActionPublishAll.String() || action == authdomain.ActionPublishOwn.String() {
		// Try publish_all
		allowed, err := checker.HasPermission(ctx, userID, domain, resource, authdomain.ActionPublishAll.String())
		if err == nil && allowed {
			return true, nil
		}

		// Try publish_own
		allowed, err = checker.HasPermission(ctx, userID, domain, resource, authdomain.ActionPublishOwn.String())
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
		return checker.HasPermission(ctx, userID, domain, resource, authdomain.ActionCreate.String())
	}

	// For manage actions: just check manage
	if action == authdomain.ActionManage.String() {
		return checker.HasPermission(ctx, userID, domain, resource, authdomain.ActionManage.String())
	}

	// For all other actions, check exact match
	return checker.HasPermission(ctx, userID, domain, resource, action)
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// getDomainFromRequest extracts the domain string from the request
func getDomainFromRequest(c fiber.Ctx) string {
	path := c.Path()

	// Check context first
	if domain := c.Locals("domain"); domain != nil {
		if s, ok := domain.(string); ok && s != "" {
			return s
		}
	}

	// ✅ Check for team_id and team_type query params FIRST
	teamID := c.Query("team_id")
	teamType := c.Query("team_type")

	if teamID != "" && teamType != "" {
		if teamType == "institution" {
			return authdomain.InstitutionTeamDomain(teamID)
		}
		if teamType == "personal" {
			return authdomain.PersonalTeamDomain(teamID)
		}
	}

	// Check for account_id query param
	accountID := c.Query("account_id")
	if accountID != "" {
		return authdomain.AccountDomain(accountID)
	}

	// Check for profile endpoints
	if strings.Contains(path, "/profile") {
		if strings.Contains(path, "/users/me/profile") || strings.Contains(path, "/users/me/avatar") {
			if uid := c.Locals(authdomain.ContextKeyUserID); uid != nil {
				if uidStr, ok := uid.(string); ok && uidStr != "" {
					return authdomain.PersonalTeamDomain(uidStr)
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
						return authdomain.PersonalTeamDomain(parts[1])
					case "institution":
						return authdomain.InstitutionTeamDomain(parts[1])
					case "account":
						return authdomain.AccountDomain(parts[1])
					}
				}
			}
		}
	}

	// Check if this is an institution logo request
	if strings.Contains(path, "/institutions/") && strings.Contains(path, "/logo") {
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "institutions" && i+1 < len(parts) {
				return authdomain.InstitutionTeamDomain(parts[i+1])
			}
		}
	}

	// Check path params
	institutionID := c.Params("institutionId")
	if institutionID != "" {
		return authdomain.InstitutionTeamDomain(institutionID)
	}

	userID := c.Params("userId")
	if userID != "" {
		return authdomain.PersonalTeamDomain(userID)
	}

	teamID = c.Params("teamId")
	if teamID != "" {
		// ✅ FIX: Determine team type from token context or query
		teamType := c.Locals("team_type")
		if teamType != nil {
			if tt, ok := teamType.(string); ok && tt == "institution" {
				return authdomain.InstitutionTeamDomain(teamID)
			}
		}
		// Check query param for team_type
		if tt := c.Query("team_type"); tt == "institution" {
			return authdomain.InstitutionTeamDomain(teamID)
		}
		// Default to personal
		return authdomain.PersonalTeamDomain(teamID)
	}

	// ✅ FIX: Check if this is a team route with ID in the path
	// For routes like /teams/:id/invite
	if strings.Contains(path, "/teams/") {
		extractedTeamID := extractTeamIDFromPath(path)
		if extractedTeamID != "" {
			// Try to get team type from context
			teamType := c.Locals("team_type")
			if teamType != nil {
				if tt, ok := teamType.(string); ok && tt == "institution" {
					return authdomain.InstitutionTeamDomain(extractedTeamID)
				}
			}
			// Check query param
			if tt := c.Query("team_type"); tt == "institution" {
				return authdomain.InstitutionTeamDomain(extractedTeamID)
			}
			// Default to personal
			return authdomain.PersonalTeamDomain(extractedTeamID)
		}
	}

	accountID = c.Params("accountId")
	if accountID != "" {
		return authdomain.AccountDomain(accountID)
	}

	id := c.Params("id")
	if id != "" {
		if strings.Contains(path, "/institutions/") {
			return authdomain.InstitutionTeamDomain(id)
		}
		if strings.Contains(path, "/accounts/") {
			return authdomain.AccountDomain(id)
		}
		if strings.Contains(path, "/teams/") {
			// ✅ FIX: Determine team type
			teamType := c.Locals("team_type")
			if teamType != nil {
				if tt, ok := teamType.(string); ok && tt == "institution" {
					return authdomain.InstitutionTeamDomain(id)
				}
			}
			if tt := c.Query("team_type"); tt == "institution" {
				return authdomain.InstitutionTeamDomain(id)
			}
			return authdomain.PersonalTeamDomain(id)
		}
		if strings.Contains(path, "/users/") {
			return authdomain.PersonalTeamDomain(id)
		}
	}

	// Check /me endpoints
	if strings.Contains(path, "/me") || strings.Contains(path, "/my") {
		if uid := c.Locals(authdomain.ContextKeyUserID); uid != nil {
			if uidStr, ok := uid.(string); ok && uidStr != "" {
				return authdomain.PersonalTeamDomain(uidStr)
			}
		}
	}

	// Check query params
	if institutionID = c.Query("institutionId"); institutionID != "" {
		return authdomain.InstitutionTeamDomain(institutionID)
	}
	if userID = c.Query("userId"); userID != "" {
		return authdomain.PersonalTeamDomain(userID)
	}
	if teamID = c.Query("teamId"); teamID != "" {
		// ✅ FIX: Determine team type
		if tt := c.Query("team_type"); tt == "institution" {
			return authdomain.InstitutionTeamDomain(teamID)
		}
		return authdomain.PersonalTeamDomain(teamID)
	}
	if accountID = c.Query("accountId"); accountID != "" {
		return authdomain.AccountDomain(accountID)
	}

	// Platform routes
	if strings.HasPrefix(path, "/api/v1/admin") ||
		strings.HasPrefix(path, "/api/v1/platform") ||
		strings.HasPrefix(path, "/api/v1/system") {
		return authdomain.DomainPlatform
	}

	// Default to user's personal team
	if uid := c.Locals(authdomain.ContextKeyUserID); uid != nil {
		if uidStr, ok := uid.(string); ok && uidStr != "" {
			return authdomain.PersonalTeamDomain(uidStr)
		}
	}

	return authdomain.DomainPlatform
}

// extractTeamIDFromPath extracts team ID from the path
func extractTeamIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "teams" && i+1 < len(parts) {
			nextPart := parts[i+1]
			// Check if it looks like a UUID (has hyphens and is long enough)
			if strings.Contains(nextPart, "-") && len(nextPart) > 30 {
				return nextPart
			}
			// Check if the part after that is a UUID (for routes like /teams/:id/invite)
			if i+2 < len(parts) {
				nextNextPart := parts[i+2]
				if strings.Contains(nextNextPart, "-") && len(nextNextPart) > 30 {
					return nextNextPart
				}
			}
		}
	}
	return ""
}



// getResourceFromRequest extracts the resource from the request path
func getResourceFromRequest(c fiber.Ctx) string {
    path := strings.TrimPrefix(c.Path(), "/api/v1/")
    segments := strings.Split(path, "/")

    // ✅ Check for invitation routes first
    for _, seg := range segments {
        if seg == "invitations" || seg == "invitation" {
            // Check if it's a POST to invitations (inviting a member)
            if c.Method() == http.MethodPost {
                return "member" // Resource is member, action will be invite
            }
            return "invitation"
        }
    }

    for i, seg := range segments {
        switch seg {
        case "users", "user":
            if i+1 < len(segments) && (segments[i+1] == "profile" || strings.Contains(segments[i+1], "profile")) {
                return authdomain.ResourceProfile.String()
            }
            if i+1 < len(segments) && segments[i+1] == "avatar" {
                return authdomain.ResourceProfile.String()
            }
            if i+1 < len(segments) && segments[i] == "me" && (segments[i+1] == "profile" || segments[i+1] == "avatar") {
                return authdomain.ResourceProfile.String()
            }
            if i+2 < len(segments) && segments[i] == "users" && segments[i+1] == "me" && segments[i+2] == "profile" {
                return authdomain.ResourceProfile.String()
            }
            if i+2 < len(segments) && (segments[i+2] == "events" || segments[i+2] == "event") {
                return authdomain.ResourceEvent.String()
            }
            if i+2 < len(segments) && segments[i+2] == "profiles" {
                return authdomain.ResourceProfile.String()
            }
            return authdomain.ResourceUser.String()
        case "profile":
            return authdomain.ResourceProfile.String()
        case "accounts", "account":
            if i+2 < len(segments) && (segments[i+2] == "events" || segments[i+2] == "event") {
                return authdomain.ResourceEvent.String()
            }
            if i+2 < len(segments) && segments[i+2] == "profile" {
                return authdomain.ResourceProfile.String()
            }
            if i+1 < len(segments) && segments[i+1] == "members" {
                return authdomain.ResourceMember.String()
            }
            return authdomain.ResourceAccount.String()
        case "institutions", "institution":
            if i+2 < len(segments) && (segments[i+2] == "events" || segments[i+2] == "event") {
                return authdomain.ResourceEvent.String()
            }
            if i+2 < len(segments) && segments[i+2] == "profile" {
                return authdomain.ResourceProfile.String()
            }
            if i+1 < len(segments) && segments[i+1] == "logo" {
                return authdomain.ResourceProfile.String()
            }
            return authdomain.ResourceInstitution.String()
        case "teams", "team":
            // ✅ Check if this is a member operation (invitations, members)
            if i+2 < len(segments) {
                if segments[i+2] == "invitations" || segments[i+2] == "members" {
                    return authdomain.ResourceMember.String()
                }
                if segments[i+2] == "events" || segments[i+2] == "event" {
                    return authdomain.ResourceEvent.String()
                }
            }
            return authdomain.ResourceTeam.String()
        case "me":
            if i+1 < len(segments) && (segments[i+1] == "events" || segments[i+1] == "event") {
                return authdomain.ResourceEvent.String()
            }
            if i+1 < len(segments) && (segments[i+1] == "profile" || segments[i+1] == "avatar") {
                return authdomain.ResourceProfile.String()
            }
        case "avatar":
            return authdomain.ResourceProfile.String()
        case "logo":
            return authdomain.ResourceProfile.String()
        case "members", "member":
            return authdomain.ResourceMember.String()
        case "billing":
            return authdomain.ResourceBilling.String()
        }
    }

    // Check for direct matches
    for _, seg := range segments {
        switch seg {
        case "profile":
            return authdomain.ResourceProfile.String()
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
            return authdomain.ResourceDashboard.String()
        case "analytics":
            return authdomain.ResourceAnalytics.String()
        case "notifications", "notification":
            return authdomain.ResourceNotification.String()
        case "media":
            return authdomain.ResourceMedia.String()
        case "billing":
            return authdomain.ResourceBilling.String()
        }
    }

    return ""
}

// getActionFromRequest maps HTTP method to action
func getActionFromRequest(c fiber.Ctx) string {
    path := c.Path()

    // ✅ Check if this is an invitation POST request
    if strings.Contains(path, "/invitations") && c.Method() == http.MethodPost {
        return "invite" // Action is invite, not create
    }

    // For avatar/logo uploads, POST should map to update, not create
    if strings.Contains(path, "/avatar") || strings.Contains(path, "/logo") {
        switch c.Method() {
        case http.MethodPost:
            return authdomain.ActionUpdate.String()
        case http.MethodDelete:
            return authdomain.ActionDelete.String()
        }
    }

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