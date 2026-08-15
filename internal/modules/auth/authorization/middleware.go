// internal/modules/auth/authorization/middleware.go

package authorization

import (
	"net/http"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"

	"github.com/gofiber/fiber/v3"
)

// AuthorizationMiddleware creates a Fiber middleware for authorization using Casbin
func AuthorizationMiddleware(enforcer *Enforcer) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user ID from context (set by auth middleware)
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

		// Determine resource from request path
		resource := getResourceFromRequest(c)

		// Determine action from HTTP method
		action := getActionFromRequest(c)

		// Get user's roles for this domain for context
		roles := enforcer.GetRolesForUserInDomain(userIDStr, domain)
		c.Locals(authdomain.ContextKeyUserRoles, roles)

		// If domain is an account domain, store account ID
		if authdomain.IsAccountDomain(domain) {
			accountID := authdomain.ExtractAccountID(domain)
			c.Locals(authdomain.ContextKeyAccountID, accountID)
		}

		// Enforce permission
		allowed, err := enforcer.Enforce(userIDStr, domain, resource, action)
		if err != nil {
			return response.InternalError(c, "Authorization error", fiber.Map{
				"error":    err.Error(),
				"user":     userIDStr,
				"domain":   domain,
				"resource": resource,
				"action":   action,
			})
		}

		if !allowed {
			return response.Forbidden(c, "Insufficient permissions", fiber.Map{
				"user":     userIDStr,
				"domain":   domain,
				"resource": resource,
				"action":   action,
				"roles":    roles,
			})
		}

		// Store domain in context for downstream handlers
		c.Locals(authdomain.ContextKeyDomain, domain)

		return c.Next()
	}
}

// RequireRoles creates middleware that requires specific roles
func RequireRoles(enforcer *Enforcer, roles ...authdomain.Role) fiber.Handler {
	return func(c fiber.Ctx) error {
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

		// Get domain from context
		domain := c.Locals(authdomain.ContextKeyDomain)
		if domain == nil {
			domain = authdomain.DomainPlatform
		}
		domainStr, ok := domain.(string)
		if !ok {
			return response.InternalError(c, "Invalid domain type", fiber.Map{
				"reason": "domain is not a string",
			})
		}

		// Get user's roles in this domain
		userRoles := enforcer.GetRolesForUserInDomain(userIDStr, domainStr)

		// Check if user has any of the required roles
		roleMap := make(map[string]bool)
		for _, r := range userRoles {
			roleMap[r] = true
		}

		// Check for each required role
		for _, required := range roles {
			if roleMap[required.String()] {
				return c.Next()
			}
		}

		// Build list of required role names for error message
		requiredRoleNames := make([]string, len(roles))
		for i, r := range roles {
			requiredRoleNames[i] = r.String()
		}

		return response.Forbidden(c, "Insufficient roles", fiber.Map{
			"user":           userIDStr,
			"domain":         domainStr,
			"required_roles": requiredRoleNames,
			"user_roles":     userRoles,
		})
	}
}

// RequireAccountAdmin creates middleware to check if user is an account admin
func RequireAccountAdmin(enforcer *Enforcer) fiber.Handler {
	return func(c fiber.Ctx) error {
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

		// Get account ID from path
		accountID := c.Params("accountId")
		if accountID == "" {
			accountID = c.Params("id")
		}

		if accountID == "" {
			accountID = c.Query("accountId")
		}

		if accountID == "" {
			return response.BadRequest(c, "Account ID required", fiber.Map{
				"reason": "accountId not found in path or query",
			})
		}

		// Check if user has account_admin role
		if !enforcer.IsAccountAdmin(accountID, userIDStr) {
			return response.Forbidden(c, "Not an account admin", fiber.Map{
				"user":          userIDStr,
				"account_id":    accountID,
				"required_role": authdomain.RoleAccountAdmin.String(),
			})
		}

		c.Locals(authdomain.ContextKeyAccountID, accountID)
		return c.Next()
	}
}

// RequireAccountRole creates middleware that checks if user has ANY account role
func RequireAccountRole(enforcer *Enforcer) fiber.Handler {
	return func(c fiber.Ctx) error {
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

		// Get account ID from path
		accountID := c.Params("accountId")
		if accountID == "" {
			accountID = c.Params("id")
		}

		if accountID == "" {
			accountID = c.Query("accountId")
		}

		if accountID == "" {
			return response.BadRequest(c, "Account ID required", fiber.Map{
				"reason": "accountId not found in path or query",
			})
		}

		domain := authdomain.AccountDomain(accountID)
		roles := enforcer.GetRolesForUserInDomain(userIDStr, domain)

		// Check if user has any account role
		hasAccess := false
		for _, role := range roles {
			switch role {
			case authdomain.RoleAccountAdmin.String(),
				authdomain.RoleEventManager.String(),
				authdomain.RoleTeamMember.String():
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			return response.Forbidden(c, "No account access", fiber.Map{
				"user":       userIDStr,
				"account_id": accountID,
				"user_roles": roles,
			})
		}

		c.Locals(authdomain.ContextKeyAccountID, accountID)
		c.Locals(authdomain.ContextKeyDomain, domain)
		return c.Next()
	}
}

// RequirePlatformRole creates middleware that requires a platform-level role
func RequirePlatformRole(enforcer *Enforcer, roles ...authdomain.Role) fiber.Handler {
	return func(c fiber.Ctx) error {
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

		// Get user's platform roles
		userRoles := enforcer.GetUserPlatformRoles(userIDStr)

		// Check if user has any of the required roles
		roleMap := make(map[string]bool)
		for _, r := range userRoles {
			roleMap[r] = true
		}

		for _, required := range roles {
			if roleMap[required.String()] {
				return c.Next()
			}
		}

		requiredRoleNames := make([]string, len(roles))
		for i, r := range roles {
			requiredRoleNames[i] = r.String()
		}

		return response.Forbidden(c, "Insufficient platform roles", fiber.Map{
			"user":           userIDStr,
			"required_roles": requiredRoleNames,
			"user_roles":     userRoles,
		})
	}
}

// RequireAccountAccess creates middleware that checks if user has ANY account role
// Use this for /accounts/me and /accounts/my endpoints
func RequireAccountAccess(enforcer *Enforcer) fiber.Handler {
	return func(c fiber.Ctx) error {
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

		// Check if user has ANY account role
		hasAccountAccess := enforcer.HasAnyAccountRole(userIDStr)

		if !hasAccountAccess {
			accounts := enforcer.GetUserAccountIDs(userIDStr)
			platformRoles := enforcer.GetUserPlatformRoles(userIDStr)

			return response.Forbidden(c, "User does not belong to any account", fiber.Map{
				"user":           userIDStr,
				"accounts":       accounts,
				"platform_roles": platformRoles,
				"message":        "User must be a member of at least one account to access this endpoint",
			})
		}

		// Get user's accounts and store the first one as default if needed
		accounts := enforcer.GetUserAccountIDs(userIDStr)
		if len(accounts) > 0 {
			c.Locals(authdomain.ContextKeyAccountID, accounts[0])
		}

		return c.Next()
	}
}

// ================================================
// HELPER FUNCTIONS
// ================================================

// getDomainFromRequest extracts the domain from the request
func getDomainFromRequest(c fiber.Ctx) string {
	// Check if domain is in context (set by previous middleware)
	if domain := c.Locals(authdomain.ContextKeyDomain); domain != nil {
		if domainStr, ok := domain.(string); ok {
			return domainStr
		}
	}

	path := c.Path()

	// Check account routes FIRST
	if strings.Contains(path, "/accounts/me") ||
		strings.Contains(path, "/account/me") ||
		strings.Contains(path, "/accounts/my") ||
		strings.Contains(path, "/account/my") {
		return authdomain.DomainPlatform
	}

	// Check for account ID in path
	accountID := c.Params("accountId")
	if accountID == "" {
		accountID = c.Params("id")
	}
	if accountID != "" {
		return authdomain.AccountDomain(accountID)
	}

	// Check if account ID is in query
	accountID = c.Query("accountId")
	if accountID != "" {
		return authdomain.AccountDomain(accountID)
	}

	// Check if it's a user-specific request (backward compatibility)
	userID := c.Params("userId")
	if userID != "" {
		return authdomain.AccountDomain(userID)
	}

	// Check if it's the current user profile
	if strings.Contains(path, "/profile") && !strings.Contains(path, "/accounts/") {
		if userID := c.Locals(authdomain.ContextKeyUserID); userID != nil {
			if userIDStr, ok := userID.(string); ok {
				return authdomain.AccountDomain(userIDStr)
			}
		}
	}

	// Check if it's a me endpoint (but NOT account me - already handled above)
	if strings.Contains(path, "/me") && !strings.Contains(path, "/accounts/") {
		if userID := c.Locals(authdomain.ContextKeyUserID); userID != nil {
			if userIDStr, ok := userID.(string); ok {
				return authdomain.AccountDomain(userIDStr)
			}
		}
	}

	// Default to platform domain
	return authdomain.DomainPlatform
}

// getResourceFromRequest extracts the resource from the request path
func getResourceFromRequest(c fiber.Ctx) string {
	path := c.Path()

	// Remove API prefix
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.TrimPrefix(path, "/api/v1/account/")

	// Get the first segment
	segments := strings.Split(path, "/")
	if len(segments) > 0 && segments[0] != "" {
		resource := segments[0]

		// Handle account "me" endpoints
		if resource == "accounts" || resource == "account" {
			if len(segments) > 1 && (segments[1] == "me" || segments[1] == "my") {
				return authdomain.ResourceAccount.String()
			}
			for _, segment := range segments {
				if segment == "members" {
					return authdomain.ResourceMember.String()
				}
			}
		}

		// Map common URL patterns to resources
		switch resource {
		case "profile", "me":
			return authdomain.ResourceProfile.String()
		case "accounts", "account":
			return authdomain.ResourceAccount.String()
		case "institutions":
			return authdomain.ResourceInstitution.String()
		case "events":
			return authdomain.ResourceEvent.String()
		case "certificates":
			return authdomain.ResourceCertificate.String()
		case "attendees":
			return authdomain.ResourceAttendee.String()
		case "payments":
			return authdomain.ResourcePayment.String()
		case "payouts":
			return authdomain.ResourcePayout.String()
		case "dashboard":
			return authdomain.ResourceDashboard.String()
		case "analytics":
			return authdomain.ResourceAnalytics.String()
		case "notifications":
			return authdomain.ResourceNotification.String()
		case "media":
			return authdomain.ResourceMedia.String()
		default:
			return resource
		}
	}

	return ""
}

// getActionFromRequest maps HTTP method to action
func getActionFromRequest(c fiber.Ctx) string {
	method := c.Method()

	switch method {
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