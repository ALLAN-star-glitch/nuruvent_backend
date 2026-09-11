// internal/modules/auth/authorization/middleware.go

package authorization

import (
	"context"
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
		if !ok || userIDStr == "" {
			return response.Unauthorized(c, "Invalid user ID", fiber.Map{
				"reason": "user_id is not a valid string",
			})
		}

		// Determine domain from request
		domain := getDomainFromRequest(c)

		// Determine resource and action
		resource := getResourceFromRequest(c)
		action := getActionFromRequest(c)

		log.Printf("🔍 AUTHZ: user=%s, domain=%s, resource=%s, action=%s",
			userIDStr, domain, resource, action)

		// ============================================================
		// SELF-SERVICE BYPASSES
		// ============================================================
		//
		// These endpoints operate on the caller's own data. The service
		// scopes the query by the authenticated user, so a domain-scoped
		// Casbin check is neither necessary nor meaningful. The resolver
		// would otherwise fall back to the token's team context and gate
		// the operation on the wrong domain.

		if isOwnProfileRequest(c) {
			log.Printf("✅ AUTHZ BYPASS: /users/me/profile")
			c.Locals(authdomain.ContextKeyDomain, domain)
			return c.Next()
		}
		if isOwnAvatarRequest(c) {
			log.Printf("✅ AUTHZ BYPASS: /users/me/avatar")
			c.Locals(authdomain.ContextKeyDomain, domain)
			return c.Next()
		}
		if isInstitutionLogoRequest(c) {
			log.Printf("✅ AUTHZ BYPASS: institution logo request")
			c.Locals(authdomain.ContextKeyDomain, domain)
			return c.Next()
		}
		if isSelfServiceListAccounts(c) {
			log.Printf("✅ AUTHZ BYPASS: self-service account listing")
			c.Locals(authdomain.ContextKeyDomain, domain)
			return c.Next()
		}
		if isSelfServiceCreateAccount(c) {
			log.Printf("✅ AUTHZ BYPASS: self-service account creation")
			c.Locals(authdomain.ContextKeyDomain, domain)
			return c.Next()
		}

		// Store resolved domain for downstream consumers
		c.Locals(authdomain.ContextKeyDomain, domain)

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

// ============================================================
// BYPASS HELPERS
// ============================================================

func isOwnProfileRequest(c fiber.Ctx) bool {
	return strings.Contains(c.Path(), "/users/me/profile")
}

func isOwnAvatarRequest(c fiber.Ctx) bool {
	return strings.Contains(c.Path(), "/users/me/avatar")
}

func isInstitutionLogoRequest(c fiber.Ctx) bool {
	path := c.Path()
	return strings.Contains(path, "/institutions/") && strings.Contains(path, "/logo")
}

// isSelfServiceListAccounts matches GET /api/v1/accounts (the list endpoint).
//
// Rationale: the response is always the caller's own accounts — the service
// scopes the query by userID. There's no cross-resource boundary to gate.
// Without this bypass, the domain resolver falls back to the token's team
// context (e.g. "institution:team:b93e58f9-...") and Casbin finds no
// account-scoped policy there, producing a spurious 403.
func isSelfServiceListAccounts(c fiber.Ctx) bool {
	if c.Method() != http.MethodGet {
		return false
	}
	path := strings.TrimSuffix(c.Path(), "/")
	return path == "/api/v1/accounts"
}

// isSelfServiceCreateAccount matches POST /api/v1/accounts/personal and
// POST /api/v1/accounts/institution.
//
// Rationale: a user creating their own account is a self-service action.
// The service handles account creation; there's no existing account to
// authorize against. Gating this on Casbin would require a chicken-and-egg
// policy — the account doesn't exist yet.
func isSelfServiceCreateAccount(c fiber.Ctx) bool {
	if c.Method() != http.MethodPost {
		return false
	}
	path := strings.TrimSuffix(c.Path(), "/")
	return path == "/api/v1/accounts/personal" || path == "/api/v1/accounts/institution"
}

// ============================================================
// PERMISSION CHECKS
// ============================================================

// checkPermissionWithFallback applies the read/update/delete fallback chains,
// then falls through to an exact match for other actions.
func checkPermissionWithFallback(
	c fiber.Ctx,
	checker authdomain.PermissionChecker,
	userID, domain, resource, action string,
) (bool, error) {
	ctx := c.Context()

	switch action {
	case authdomain.ActionRead.String():
		return checkAny(
			ctx, checker, userID, domain, resource,
			[]string{
				authdomain.ActionRead.String(),
				authdomain.ActionReadAll.String(),
				authdomain.ActionReadOwn.String(),
			},
		)

	case authdomain.ActionUpdate.String():
		return checkAny(
			ctx, checker, userID, domain, resource,
			[]string{
				authdomain.ActionUpdate.String(),
				authdomain.ActionUpdateAll.String(),
				authdomain.ActionUpdateOwn.String(),
			},
		)

	case authdomain.ActionDelete.String():
		return checkAny(
			ctx, checker, userID, domain, resource,
			[]string{
				authdomain.ActionDelete.String(),
				authdomain.ActionDeleteAll.String(),
				authdomain.ActionDeleteOwn.String(),
			},
		)

	case authdomain.ActionPublishAll.String(), authdomain.ActionPublishOwn.String():
		return checkAny(
			ctx, checker, userID, domain, resource,
			[]string{
				authdomain.ActionPublishAll.String(),
				authdomain.ActionPublishOwn.String(),
			},
		)
	}

	// Exact match for create, manage, and any other action
	return checker.HasPermission(ctx, userID, domain, resource, action)
}

// checkAny returns (true, nil) on the first granted permission,
// (false, err) on the first permission-check error, or (false, nil) if none match.
func checkAny(
	ctx context.Context,
	checker authdomain.PermissionChecker,
	userID, domain, resource string,
	actions []string,
) (bool, error) {
	for _, action := range actions {
		allowed, err := checker.HasPermission(ctx, userID, domain, resource, action)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}

// ============================================================
// DOMAIN RESOLUTION
// ============================================================

// getDomainFromRequest resolves the Casbin domain for a request.
//
// Priority order:
//  1. Explicit override via c.Locals(ContextKeyDomain) — set by upstream middleware
//  2. Query param override — ?team_id=X&team_type=Y (admin cross-team)
//  3. Query param override — ?account_id=X
//  4. Token team context (team_id + team_type from JWT) — default for auth'd users
//  5. Profile-specific paths (/users/me/profile, /users/me/avatar)
//  6. Path params (:institutionId, :userId, :teamId, :accountId, :id)
//  7. /teams/... path heuristic
//  8. /me and /my endpoints (personal scope)
//  9. Platform routes (/admin, /platform, /system)
// 10. Fallback: personal team of the authenticated user
func getDomainFromRequest(c fiber.Ctx) string {
	path := c.Path()

	// 1. Explicit override
	if domain := c.Locals(authdomain.ContextKeyDomain); domain != nil {
		if s, ok := domain.(string); ok && s != "" {
			return s
		}
	}

	// 2. Team override via query
	if teamID, teamType := c.Query("team_id"), c.Query("team_type"); teamID != "" {
		return authdomain.BuildTeamDomain(teamType, teamID)
	}

	// 3. Account override via query
	if accountID := c.Query("account_id"); accountID != "" {
		return authdomain.AccountDomain(accountID)
	}

	// 4. Token team context — default for authenticated requests
	if domain := domainFromToken(c); domain != "" {
		return domain
	}

	// 5. Profile paths
	if strings.Contains(path, "/users/me/profile") || strings.Contains(path, "/users/me/avatar") {
		if domain := personalDomainForCurrentUser(c); domain != "" {
			return domain
		}
	}
	if strings.Contains(path, "/profile/organizer") {
		if domain := domainFromScopeQuery(c); domain != "" {
			return domain
		}
	}

	// 6. Institution logo path
	if strings.Contains(path, "/institutions/") && strings.Contains(path, "/logo") {
		if domain := institutionDomainFromPath(path); domain != "" {
			return domain
		}
	}

	// 7. Path params
	if domain := domainFromPathParams(c, path); domain != "" {
		return domain
	}

	// 8. /teams/... path heuristic
	if strings.Contains(path, "/teams/") {
		if domain := domainFromTeamsPath(c, path); domain != "" {
			return domain
		}
	}

	// 9. /me and /my endpoints (personal scope)
	if strings.Contains(path, "/me") || strings.Contains(path, "/my") {
		if domain := personalDomainForCurrentUser(c); domain != "" {
			return domain
		}
	}

	// 10. Platform routes
	if strings.HasPrefix(path, "/api/v1/admin") ||
		strings.HasPrefix(path, "/api/v1/platform") ||
		strings.HasPrefix(path, "/api/v1/system") {
		return authdomain.DomainPlatform
	}

	// 11. Fallback: personal team of the authenticated user
	if domain := personalDomainForCurrentUser(c); domain != "" {
		return domain
	}

	return authdomain.DomainPlatform
}

// domainFromToken reads team_id + team_type from Locals and returns the domain.
// Returns "" if no team context is present.
func domainFromToken(c fiber.Ctx) string {
	tidRaw := c.Locals(authdomain.ContextKeyTeamID)
	if tidRaw == nil {
		return ""
	}
	tid, ok := tidRaw.(string)
	if !ok || tid == "" {
		return ""
	}

	teamType := authdomain.TeamTypePersonal // safe default
	if ttRaw := c.Locals(authdomain.ContextKeyTeamType); ttRaw != nil {
		if tt, ok := ttRaw.(string); ok && tt != "" {
			teamType = tt
		}
	}

	return authdomain.BuildTeamDomain(teamType, tid)
}

// personalDomainForCurrentUser returns personal:team:<user_id> for the auth'd user.
func personalDomainForCurrentUser(c fiber.Ctx) string {
	uidRaw := c.Locals(authdomain.ContextKeyUserID)
	if uidRaw == nil {
		return ""
	}
	uid, ok := uidRaw.(string)
	if !ok || uid == "" {
		return ""
	}
	return authdomain.PersonalTeamDomain(uid)
}

// domainFromScopeQuery parses ?scope=<type>:<id> (used by /profile/organizer).
func domainFromScopeQuery(c fiber.Ctx) string {
	scope := c.Query("scope")
	if scope == "" {
		return ""
	}
	parts := strings.SplitN(scope, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	switch parts[0] {
	case "personal":
		return authdomain.PersonalTeamDomain(parts[1])
	case "institution":
		return authdomain.InstitutionTeamDomain(parts[1])
	case "account":
		return authdomain.AccountDomain(parts[1])
	}
	return ""
}

// institutionDomainFromPath extracts the institution ID from /institutions/:id/logo-style paths.
func institutionDomainFromPath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if p == "institutions" && i+1 < len(parts) {
			return authdomain.InstitutionTeamDomain(parts[i+1])
		}
	}
	return ""
}

// domainFromPathParams resolves the domain from route params, in priority order.
func domainFromPathParams(c fiber.Ctx, path string) string {
	if id := c.Params("institutionId"); id != "" {
		return authdomain.InstitutionTeamDomain(id)
	}
	if id := c.Params("userId"); id != "" {
		return authdomain.PersonalTeamDomain(id)
	}
	if id := c.Params("teamId"); id != "" {
		return teamDomainWithType(c, id)
	}
	if id := c.Params("accountId"); id != "" {
		return authdomain.AccountDomain(id)
	}
	if id := c.Params("id"); id != "" {
		switch {
		case strings.Contains(path, "/institutions/"):
			return authdomain.InstitutionTeamDomain(id)
		case strings.Contains(path, "/accounts/"):
			return authdomain.AccountDomain(id)
		case strings.Contains(path, "/teams/"):
			return teamDomainWithType(c, id)
		case strings.Contains(path, "/users/"):
			return authdomain.PersonalTeamDomain(id)
		}
	}
	return ""
}

// teamDomainWithType picks institution vs personal based on token/query context.
// Defaults to institution when no explicit type is available.
func teamDomainWithType(c fiber.Ctx, teamID string) string {
	if tt := tokenTeamType(c); tt != "" {
		return authdomain.BuildTeamDomain(tt, teamID)
	}
	if tt := c.Query("team_type"); tt != "" {
		return authdomain.BuildTeamDomain(tt, teamID)
	}
	return authdomain.InstitutionTeamDomain(teamID)
}

// domainFromTeamsPath handles routes like /teams/:id/... where the ID is a UUID.
func domainFromTeamsPath(c fiber.Ctx, path string) string {
	teamID := extractTeamIDFromPath(path)
	if teamID == "" {
		return ""
	}
	return teamDomainWithType(c, teamID)
}

// tokenTeamType returns the team_type stored by AuthMiddleware, or "" if absent.
func tokenTeamType(c fiber.Ctx) string {
	raw := c.Locals(authdomain.ContextKeyTeamType)
	if raw == nil {
		return ""
	}
	if s, ok := raw.(string); ok {
		return s
	}
	return ""
}

// extractTeamIDFromPath returns the UUID after /teams/ in the path.
func extractTeamIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if p == "teams" && i+1 < len(parts) {
			next := parts[i+1]
			if strings.Contains(next, "-") && len(next) > 30 {
				return next
			}
			if i+2 < len(parts) {
				nextNext := parts[i+2]
				if strings.Contains(nextNext, "-") && len(nextNext) > 30 {
					return nextNext
				}
			}
		}
	}
	return ""
}

// ============================================================
// RESOURCE RESOLUTION
// ============================================================

// getResourceFromRequest extracts the Casbin resource from the request path.
func getResourceFromRequest(c fiber.Ctx) string {
	path := strings.TrimPrefix(c.Path(), "/api/v1/")
	segments := strings.Split(path, "/")

	// Special-case: invitation POSTs are member/invite actions
	for _, seg := range segments {
		if seg == "invitations" || seg == "invitation" {
			return authdomain.ResourceMember.String()
		}
	}

	// Contextual resource resolution (path-position-aware)
	for i, seg := range segments {
		switch seg {
		case "users", "user":
			if r := userSubResource(segments, i); r != "" {
				return r
			}
			return authdomain.ResourceUser.String()
		case "accounts", "account":
			if r := accountSubResource(segments, i); r != "" {
				return r
			}
			return authdomain.ResourceAccount.String()
		case "institutions", "institution":
			if r := institutionSubResource(segments, i); r != "" {
				return r
			}
			return authdomain.ResourceInstitution.String()
		case "teams", "team":
			if r := teamSubResource(segments, i); r != "" {
				return r
			}
			return authdomain.ResourceTeam.String()
		case "profile":
			return authdomain.ResourceProfile.String()
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

	// Direct resource match
	for _, seg := range segments {
		switch seg {
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
		}
	}

	return ""
}

// ---- Resource sub-resolvers (keep each path family in one place) ----

func userSubResource(segments []string, i int) string {
	if i+1 >= len(segments) {
		return ""
	}
	next := segments[i+1]
	switch {
	case strings.Contains(next, "profile"):
		return authdomain.ResourceProfile.String()
	case next == "avatar":
		return authdomain.ResourceProfile.String()
	case i+2 < len(segments) && (segments[i+2] == "events" || segments[i+2] == "event"):
		return authdomain.ResourceEvent.String()
	case i+2 < len(segments) && segments[i+2] == "profiles":
		return authdomain.ResourceProfile.String()
	}
	return ""
}

func accountSubResource(segments []string, i int) string {
	if i+1 < len(segments) && segments[i+1] == "members" {
		return authdomain.ResourceMember.String()
	}
	if i+2 < len(segments) {
		switch segments[i+2] {
		case "events", "event":
			return authdomain.ResourceEvent.String()
		case "profile":
			return authdomain.ResourceProfile.String()
		}
	}
	return ""
}

func institutionSubResource(segments []string, i int) string {
	if i+1 < len(segments) && segments[i+1] == "logo" {
		return authdomain.ResourceProfile.String()
	}
	if i+2 < len(segments) {
		switch segments[i+2] {
		case "events", "event":
			return authdomain.ResourceEvent.String()
		case "profile":
			return authdomain.ResourceProfile.String()
		}
	}
	return ""
}

func teamSubResource(segments []string, i int) string {
	if i+2 >= len(segments) {
		return ""
	}
	switch segments[i+2] {
	case "invitations", "members":
		return authdomain.ResourceMember.String()
	case "events", "event":
		return authdomain.ResourceEvent.String()
	}
	return ""
}

// ============================================================
// ACTION RESOLUTION
// ============================================================

// getActionFromRequest maps HTTP method to Casbin action.
func getActionFromRequest(c fiber.Ctx) string {
	path := c.Path()

	// POST /invitations → invite (not create)
	if strings.Contains(path, "/invitations") && c.Method() == http.MethodPost {
		return authdomain.ActionInvite.String()
	}

	// POST/DELETE on avatar or logo → update/delete (not create)
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