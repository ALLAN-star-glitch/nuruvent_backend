package authorization

import (
	"net/http"
	"strings"

	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/response"
	"github.com/gofiber/fiber/v3"
)

// AuthorizationMiddleware creates a Fiber middleware for authorization using Casbin
func AuthorizationMiddleware(enforcer *Enforcer) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user ID from context (set by auth middleware)
		userID := c.Locals("user_id")
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
		c.Locals(ContextKeyUserRoles, roles)

		// If domain is a business domain, store business ID
		if IsBusinessDomain(domain) {
			businessID := ExtractBusinessID(domain)
			c.Locals(ContextKeyBusinessID, businessID)
		}

		// Enforce permission
		allowed, err := enforcer.EnforceWithContext(userIDStr, domain, resource, action)
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
		c.Locals(ContextKeyDomain, domain)

		return c.Next()
	}
}

// RequireRoles creates middleware that requires specific roles
func RequireRoles(enforcer *Enforcer, roles ...Role) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Locals("user_id")
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
		domain := c.Locals(ContextKeyDomain)
		if domain == nil {
			domain = DomainPlatform
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

// RequireBusinessAdmin creates middleware to check if user is a business admin
func RequireBusinessAdmin(enforcer *Enforcer) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Locals("user_id")
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

		// Get business ID from path
		businessID := c.Params("businessId")
		if businessID == "" {
			businessID = c.Params("id")
		}

		if businessID == "" {
			businessID = c.Query("businessId")
		}

		if businessID == "" {
			return response.BadRequest(c, "Business ID required", fiber.Map{
				"reason": "businessId not found in path or query",
			})
		}

		// Check if user has business_admin role
		if !enforcer.IsBusinessAdmin(userIDStr, businessID) {
			return response.Forbidden(c, "Not a business admin", fiber.Map{
				"user":          userIDStr,
				"business_id":   businessID,
				"required_role": RoleBusinessAdmin.String(),
			})
		}

		c.Locals(ContextKeyBusinessID, businessID)
		return c.Next()
	}
}

// RequireBusinessRole creates middleware to check if user has ANY business role
func RequireBusinessRole(enforcer *Enforcer) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Locals("user_id")
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

		// Get business ID from path
		businessID := c.Params("businessId")
		if businessID == "" {
			businessID = c.Params("id")
		}

		if businessID == "" {
			businessID = c.Query("businessId")
		}

		if businessID == "" {
			return response.BadRequest(c, "Business ID required", fiber.Map{
				"reason": "businessId not found in path or query",
			})
		}

		domain := BusinessDomain(businessID)
		roles := enforcer.GetRolesForUserInDomain(userIDStr, domain)

		// Check if user has any business role
		hasAccess := false
		for _, role := range roles {
			switch role {
			case RoleBusinessAdmin.String(),
				RoleEventManager.String(),
				RoleMember.String():
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			return response.Forbidden(c, "No business access", fiber.Map{
				"user":        userIDStr,
				"business_id": businessID,
				"user_roles":  roles,
			})
		}

		c.Locals(ContextKeyBusinessID, businessID)
		c.Locals(ContextKeyDomain, domain)
		return c.Next()
	}
}

// RequirePlatformRole creates middleware that requires a platform-level role
func RequirePlatformRole(enforcer *Enforcer, roles ...Role) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Locals("user_id")
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

// RequireBusinessAccess creates middleware that checks if user has ANY business role
// Use this for /businesses/me and /businesses/my endpoints
func RequireBusinessAccess(enforcer *Enforcer) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Locals("user_id")
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

		// Check if user has ANY business role
		hasBusinessAccess := enforcer.HasAnyBusinessRole(userIDStr)

		if !hasBusinessAccess {
			businesses := enforcer.GetUserBusinesses(userIDStr)
			platformRoles := enforcer.GetUserPlatformRoles(userIDStr)

			return response.Forbidden(c, "User does not belong to any business", fiber.Map{
				"user":           userIDStr,
				"businesses":     businesses,
				"platform_roles": platformRoles,
				"message":        "User must be a member of at least one business to access this endpoint",
			})
		}

		// Get user's businesses and store the first one as default if needed
		businesses := enforcer.GetUserBusinesses(userIDStr)
		if len(businesses) > 0 {
			c.Locals(ContextKeyBusinessID, businesses[0])
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
	if domain := c.Locals(ContextKeyDomain); domain != nil {
		if domainStr, ok := domain.(string); ok {
			return domainStr
		}
	}

	path := c.Path()

	// Check business routes FIRST (before user routes)
	if strings.Contains(path, "/businesses/me") ||
		strings.Contains(path, "/business/me") ||
		strings.Contains(path, "/businesses/my") ||
		strings.Contains(path, "/business/my") {
		return DomainPlatform
	}

	// Check for business ID in path - check BOTH "id" and "businessId"
	businessID := c.Params("businessId")
	if businessID == "" {
		businessID = c.Params("id")
	}
	if businessID != "" {
		return BusinessDomain(businessID)
	}

	// Check if business ID is in query
	businessID = c.Query("businessId")
	if businessID != "" {
		return BusinessDomain(businessID)
	}

	// Check if it's a user-specific request
	userID := c.Params("userId")
	if userID != "" {
		return UserDomain(userID)
	}

	// Check if it's the current user profile
	if strings.Contains(path, "/profile") && !strings.Contains(path, "/businesses/") {
		if userID := c.Locals("user_id"); userID != nil {
			if userIDStr, ok := userID.(string); ok {
				return UserDomain(userIDStr)
			}
		}
	}

	// Check if it's a me endpoint (but NOT business me - already handled above)
	if strings.Contains(path, "/me") && !strings.Contains(path, "/businesses/") {
		if userID := c.Locals("user_id"); userID != nil {
			if userIDStr, ok := userID.(string); ok {
				return UserDomain(userIDStr)
			}
		}
	}

	// Default to platform domain
	return DomainPlatform
}

// getResourceFromRequest extracts the resource from the request path
func getResourceFromRequest(c fiber.Ctx) string {
	path := c.Path()

	// Remove API prefix
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.TrimPrefix(path, "/api/v1/business/")

	// Get the first segment
	segments := strings.Split(path, "/")
	if len(segments) > 0 && segments[0] != "" {
		resource := segments[0]

		// Handle business "me" endpoints
		if resource == "businesses" || resource == "business" {
			if len(segments) > 1 && (segments[1] == "me" || segments[1] == "my") {
				return ResourceBusiness.String()
			}
			for _, segment := range segments {
				if segment == "members" {
					return ResourceMember.String()
				}
			}
		}

		// Map common URL patterns to resources
		switch resource {
		case "profile", "me":
			return ResourceUser.String()
		case "businesses", "business":
			return ResourceBusiness.String()
		case "events":
			return ResourceEvent.String()
		case "certificates":
			return ResourceCertificate.String()
		case "attendees":
			return ResourceAttendee.String()
		case "payments":
			return ResourcePayment.String()
		case "payouts":
			return ResourcePayout.String()
		case "dashboard":
			return ResourceDashboard.String()
		case "analytics":
			return ResourceAnalytics.String()
		case "notifications":
			return ResourceNotification.String()
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
		return ActionRead.String()
	case http.MethodPost:
		return ActionCreate.String()
	case http.MethodPut, http.MethodPatch:
		return ActionUpdate.String()
	case http.MethodDelete:
		return ActionDelete.String()
	default:
		return ActionRead.String()
	}
}