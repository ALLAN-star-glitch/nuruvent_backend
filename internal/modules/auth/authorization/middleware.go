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

		// 🔍 DEBUG LOGGING
		log.Printf("🔍 AUTHZ DEBUG: userID=%s, domain=%s, resource=%s, action=%s, path=%s", 
			userIDStr, domain, resource, action, c.Path())

		// Get user's roles for this domain for context
		roles := enforcer.GetRolesForUserInDomain(userIDStr, domain)
		log.Printf("🔍 AUTHZ DEBUG: userID=%s, domain=%s, roles=%v", userIDStr, domain, roles)

		c.Locals(authdomain.ContextKeyUserRoles, roles)

		// If domain is a team domain, store the domain
		if authdomain.IsTeamDomain(domain) {
			c.Locals(authdomain.ContextKeyDomain, domain)
		}

		// Enforce permission
		allowed, err := enforcer.Enforce(userIDStr, domain, resource, action)
		log.Printf("🔍 AUTHZ DEBUG: allowed=%v, err=%v", allowed, err)

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

// ============================================================
// PERSONAL TEAM ADMIN MIDDLEWARE
// Domain: personal:team:{user_id}
// ============================================================

// RequirePersonalTeamAdmin creates middleware to check if user is admin of their personal team
func RequirePersonalTeamAdmin(enforcer *Enforcer) fiber.Handler {
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

		// Check if user is admin of their personal team
		if !enforcer.IsPersonalTeamAdmin(userIDStr) {
			return response.Forbidden(c, "Not a personal team admin", fiber.Map{
				"user":          userIDStr,
				"required_role": authdomain.RoleAccountAdmin.String(),
			})
		}

		return c.Next()
	}
}

// ============================================================
// INSTITUTION TEAM ADMIN MIDDLEWARE
// Domain: institution:team:{institution_id}
// ============================================================

// RequireInstitutionTeamAdmin creates middleware to check if user is admin of an institution team
func RequireInstitutionTeamAdmin(enforcer *Enforcer) fiber.Handler {
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

		// Get institution ID from path
		institutionID := c.Params("institutionId")
		if institutionID == "" {
			institutionID = c.Params("id")
		}

		if institutionID == "" {
			institutionID = c.Query("institutionId")
		}

		if institutionID == "" {
			return response.BadRequest(c, "Institution ID required", fiber.Map{
				"reason": "institutionId not found in path or query",
			})
		}

		// Check if user has account_admin role for this institution
		if !enforcer.IsInstitutionTeamAdmin(userIDStr, institutionID) {
			return response.Forbidden(c, "Not an institution team admin", fiber.Map{
				"user":           userIDStr,
				"institution_id": institutionID,
				"required_role":  authdomain.RoleAccountAdmin.String(),
			})
		}

		return c.Next()
	}
}

// ============================================================
// GENERIC TEAM MIDDLEWARE (DEPRECATED)
// ============================================================

// RequireTeamAdmin is deprecated - Use RequirePersonalTeamAdmin or RequireInstitutionTeamAdmin
// Deprecated: Use RequirePersonalTeamAdmin or RequireInstitutionTeamAdmin instead
func RequireTeamAdmin(enforcer *Enforcer) fiber.Handler {
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

		// Get team ID from path
		teamID := c.Params("teamId")
		if teamID == "" {
			teamID = c.Params("id")
		}

		if teamID == "" {
			teamID = c.Query("teamId")
		}

		if teamID == "" {
			return response.BadRequest(c, "Team ID required", fiber.Map{
				"reason": "teamId not found in path or query",
			})
		}

		// Try personal team first
		if enforcer.IsPersonalTeamAdmin(userIDStr) {
			return c.Next()
		}

		// Try institution team
		if enforcer.IsInstitutionTeamAdmin(userIDStr, teamID) {
			return c.Next()
		}

		return response.Forbidden(c, "Not a team admin", fiber.Map{
			"user":          userIDStr,
			"team_id":       teamID,
			"required_role": authdomain.RoleAccountAdmin.String(),
		})
	}
}

// ============================================================
// TEAM ROLE MIDDLEWARE (DEPRECATED)
// ============================================================

// RequireTeamRole is deprecated - Use specific middleware instead
// Deprecated: Use RequirePersonalTeamAdmin, RequireInstitutionTeamAdmin, or RequirePlatformRole
func RequireTeamRole(enforcer *Enforcer) fiber.Handler {
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

		// Get team ID from path
		teamID := c.Params("teamId")
		if teamID == "" {
			teamID = c.Params("id")
		}

		if teamID == "" {
			teamID = c.Query("teamId")
		}

		if teamID == "" {
			return response.BadRequest(c, "Team ID required", fiber.Map{
				"reason": "teamId not found in path or query",
			})
		}

		// Check both personal and institution domains
		personalDomain := authdomain.PersonalTeamDomain(teamID)
		personalRoles := enforcer.GetRolesForUserInDomain(userIDStr, personalDomain)

		institutionDomain := authdomain.InstitutionTeamDomain(teamID)
		institutionRoles := enforcer.GetRolesForUserInDomain(userIDStr, institutionDomain)

		allRoles := append(personalRoles, institutionRoles...)

		hasAccess := false
		for _, role := range allRoles {
			switch role {
			case authdomain.RoleAccountAdmin.String(),
				authdomain.RoleEventManager.String(),
				authdomain.RoleTeamMember.String():
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			return response.Forbidden(c, "No team access", fiber.Map{
				"user":            userIDStr,
				"team_id":         teamID,
				"personal_roles":  personalRoles,
				"institution_roles": institutionRoles,
			})
		}

		return c.Next()
	}
}

// ============================================================
// TEAM ACCESS MIDDLEWARE
// ============================================================

// RequireTeamAccess creates middleware that checks if user has ANY team access
func RequireTeamAccess(enforcer *Enforcer) fiber.Handler {
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

		// Check if user has ANY team role
		hasTeamAccess := enforcer.HasAnyTeamRole(userIDStr)

		if !hasTeamAccess {
			personalTeams := enforcer.GetUserPersonalTeamIDs(userIDStr)
			institutionTeams := enforcer.GetUserInstitutionTeamIDs(userIDStr)
			platformRoles := enforcer.GetUserPlatformRoles(userIDStr)

			return response.Forbidden(c, "User does not belong to any team", fiber.Map{
				"user":              userIDStr,
				"personal_teams":    personalTeams,
				"institution_teams": institutionTeams,
				"platform_roles":    platformRoles,
				"message":           "User must be a member of at least one team to access this endpoint",
			})
		}

		return c.Next()
	}
}

// ============================================================
// PLATFORM ROLE MIDDLEWARE
// ============================================================

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

// ============================================================
// HELPER FUNCTIONS
// ============================================================
func getDomainFromRequest(c fiber.Ctx) string {
	// Check if domain is in context (set by previous middleware)
	if domain := c.Locals(authdomain.ContextKeyDomain); domain != nil {
		if domainStr, ok := domain.(string); ok {
			log.Printf("🔍 DOMAIN DEBUG: returning domain from context: %s", domainStr)
			return domainStr
		}
	}

	path := c.Path()

	log.Printf("🔍 DOMAIN DEBUG: path=%s", path)

	// ============================================================
	// 1. CHECK FOR ID IN PATH (MOST SPECIFIC)
	// ============================================================

	// ✅ Check for institutionId
	institutionID := c.Params("institutionId")
	if institutionID != "" {
		return authdomain.InstitutionTeamDomain(institutionID)
	}

	// ✅ Check for teamId
	teamID := c.Params("teamId")
	if teamID != "" {
		return authdomain.PersonalTeamDomain(teamID)
	}

	// ✅ ADD THIS: Check for userId
	userID := c.Params("userId")
	if userID != "" {
		return authdomain.PersonalTeamDomain(userID)
	}

	// Check for generic "id" parameter
	id := c.Params("id")
	if id != "" {
		// Check if this is an institution route
		if strings.Contains(path, "/institutions/") || strings.Contains(path, "/institution/") {
			return authdomain.InstitutionTeamDomain(id)
		}
		// Check if this is a user route
		if strings.Contains(path, "/users/") || strings.Contains(path, "/user/") {
			return authdomain.PersonalTeamDomain(id)
		}
		// Check if this is a team route
		if strings.Contains(path, "/teams/") || strings.Contains(path, "/team/") {
			return authdomain.PersonalTeamDomain(id)
		}
	}

	// ============================================================
	// 2. PARSE ID FROM URL PATH
	// ============================================================

	cleanPath := strings.TrimPrefix(path, "/api/v1/")
	cleanPath = strings.TrimPrefix(cleanPath, "/api/v1")

	segments := strings.Split(cleanPath, "/")
	for i, segment := range segments {
		if segment == "" {
			continue
		}

		// Check for institutions
		if segment == "institutions" || segment == "institution" {
			if i+1 < len(segments) && segments[i+1] != "" {
				candidate := segments[i+1]
				if len(candidate) == 36 && strings.Count(candidate, "-") == 4 {
					return authdomain.InstitutionTeamDomain(candidate)
				}
			}
		}
		// Check for users
		if segment == "users" || segment == "user" {
			if i+1 < len(segments) && segments[i+1] != "" {
				candidate := segments[i+1]
				if len(candidate) == 36 && strings.Count(candidate, "-") == 4 {
					return authdomain.PersonalTeamDomain(candidate)
				}
			}
		}
		// Check for teams
		if segment == "teams" || segment == "team" {
			if i+1 < len(segments) && segments[i+1] != "" {
				candidate := segments[i+1]
				if len(candidate) == 36 && strings.Count(candidate, "-") == 4 {
					return authdomain.PersonalTeamDomain(candidate)
				}
			}
		}
	}

	// ============================================================
	// 3. CHECK FOR ME/MY ENDPOINTS
	// ============================================================

	if strings.Contains(path, "/me") || strings.Contains(path, "/my") {
		if userID := c.Locals(authdomain.ContextKeyUserID); userID != nil {
			if userIDStr, ok := userID.(string); ok && userIDStr != "" {
				return authdomain.PersonalTeamDomain(userIDStr)
			}
		}
	}

	// ============================================================
	// 4. CHECK QUERY PARAMETERS
	// ============================================================

	institutionID = c.Query("institutionId")
	if institutionID != "" {
		return authdomain.InstitutionTeamDomain(institutionID)
	}

	userID = c.Query("userId")
	if userID != "" {
		return authdomain.PersonalTeamDomain(userID)
	}

	teamID = c.Query("teamId")
	if teamID != "" {
		return authdomain.PersonalTeamDomain(teamID)
	}

	// ============================================================
	// 5. CHECK USER PROFILE ROUTES
	// ============================================================

	if strings.Contains(path, "/profile") && !strings.Contains(path, "/institutions/") && !strings.Contains(path, "/teams/") && !strings.Contains(path, "/users/") {
		if userID := c.Locals(authdomain.ContextKeyUserID); userID != nil {
			if userIDStr, ok := userID.(string); ok {
				return authdomain.PersonalTeamDomain(userIDStr)
			}
		}
	}

	// ============================================================
	// 6. CHECK PLATFORM ROUTES
	// ============================================================

	if strings.HasPrefix(path, "/api/v1/admin") ||
		strings.HasPrefix(path, "/api/v1/platform") ||
		strings.HasPrefix(path, "/api/v1/system") {
		return authdomain.DomainPlatform
	}

	// ============================================================
	// 7. DEFAULT - Use user's personal team if available
	// ============================================================

	if userID := c.Locals(authdomain.ContextKeyUserID); userID != nil {
		if userIDStr, ok := userID.(string); ok && userIDStr != "" {
			return authdomain.PersonalTeamDomain(userIDStr)
		}
	}

	return authdomain.DomainPlatform
}

// getResourceFromRequest extracts the resource from the request path
func getResourceFromRequest(c fiber.Ctx) string {
	path := c.Path()

	// Remove API prefix
	cleanPath := strings.TrimPrefix(path, "/api/v1/")
	cleanPath = strings.TrimPrefix(cleanPath, "/api/v1")

	segments := strings.Split(cleanPath, "/")

	// ============================================================
	// 1. CHECK FOR PATTERN: /users/{userId}/events
	// ============================================================
	for i, segment := range segments {
		if segment == "users" || segment == "user" {
			// Look ahead for "events" after the user ID
			if i+2 < len(segments) {
				if segments[i+2] == "events" || segments[i+2] == "event" {
					return authdomain.ResourceEvent.String()
				}
			}
			// If no "events" after user ID, then it's a user resource
			return authdomain.ResourceUser.String()
		}

		// ============================================================
		// 2. CHECK FOR PATTERN: /institutions/{id}/events
		// ============================================================
		if segment == "institutions" || segment == "institution" {
			if i+2 < len(segments) {
				if segments[i+2] == "events" || segments[i+2] == "event" {
					return authdomain.ResourceEvent.String()
				}
			}
			return authdomain.ResourceInstitution.String()
		}

		// ============================================================
		// 3. CHECK FOR PATTERN: /teams/{id}/events
		// ============================================================
		if segment == "teams" || segment == "team" {
			if i+2 < len(segments) {
				if segments[i+2] == "events" || segments[i+2] == "event" {
					return authdomain.ResourceEvent.String()
				}
			}
			return authdomain.ResourceTeam.String()
		}

		// ============================================================
		// 4. CHECK FOR /users/me/events
		// ============================================================
		if segment == "me" && i+1 < len(segments) {
			if segments[i+1] == "events" || segments[i+1] == "event" {
				return authdomain.ResourceEvent.String()
			}
		}
	}

	// ============================================================
	// 5. CHECK FOR DIRECT RESOURCE MATCHES
	// ============================================================
	for _, segment := range segments {
		if resource := mapResource(segment); resource != "" {
			return resource
		}
	}

	return ""
}

// mapResource maps URL segment to resource constant
func mapResource(segment string) string {
	switch segment {
	case "profile":
		return authdomain.ResourceProfile.String()
	case "users", "user":
		return authdomain.ResourceUser.String()
	case "institutions", "institution":
		return authdomain.ResourceInstitution.String()
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
	case "dashboard":
		return authdomain.ResourceDashboard.String()
	case "analytics":
		return authdomain.ResourceAnalytics.String()
	case "notifications", "notification":
		return authdomain.ResourceNotification.String()
	case "media":
		return authdomain.ResourceMedia.String()
	case "members", "member":
		return authdomain.ResourceMember.String()
	case "teams", "team":
		return authdomain.ResourceTeam.String()
	case "team-types", "team-type":
		return authdomain.ResourceTeamType.String()
	default:
		return ""
	}
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