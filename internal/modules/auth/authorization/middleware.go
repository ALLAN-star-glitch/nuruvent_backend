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

// AuthorizationMiddleware enforces permissions using Casbin.
//
// DOMAIN MODEL:
//
//   "platform"       — Nuruvent staff only
//   "account:<uuid>" — every tenant account
//
// Teams are NOT authz domains. Team-scoped routes fall through to
// DomainDeferred; the service layer performs the team→account resolution
// and the authorization check.
//
// Resolution pipeline per request:
//
//   1. Extract caller identity (userID) — required.
//   2. Resolve the domain (platform, account, or deferred).
//   3. Resolve resource + action from the path/method.
//   4. Short-circuit self-service routes.
//   5. Short-circuit deferred routes — service authorizes.
//   6. Casbin check with fallback chain.
func AuthorizationMiddleware(checker authdomain.PermissionChecker) fiber.Handler {
	return func(c fiber.Ctx) error {
		// ---- 1. Identity ----
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

		// ---- 2. Resolve domain, resource, action ----
		domain := getDomainFromRequest(c)
		resource := getResourceFromRequest(c)
		action := getActionFromRequest(c)

		log.Printf("🔍 AUTHZ: user=%s, domain=%s, resource=%s, action=%s",
			userIDStr, domain, resource, action)

		// ---- 3. Self-service bypasses ----
		//
		// These endpoints operate on the caller's own data. The service
		// scopes the query by the authenticated user, so a domain-scoped
		// Casbin check is neither necessary nor meaningful.
		if isOwnProfileRequest(c) ||
			isOwnAvatarRequest(c) ||
			isInstitutionLogoRequest(c) ||
			isSelfServiceListAccounts(c) ||
			isSelfServiceCreateAccount(c) {
			log.Printf("✅ AUTHZ BYPASS: %s %s", c.Method(), c.Path())
			c.Locals(authdomain.ContextKeyDomain, domain)
			return c.Next()
		}

		// ---- 4. Deferred domain ----
		//
		// The resolver returns DomainDeferred when it cannot determine the
		// account from the URL alone (slug routes, team-scoped routes, etc.).
		// The service layer is responsible for authorization in these cases.
		if domain == authdomain.DomainDeferred {
			log.Printf("⏸  AUTHZ DEFERRED: %s %s (service must authorize)", c.Method(), c.Path())
			c.Locals(authdomain.ContextKeyDomain, domain)
			return c.Next()
		}

		// ---- 5. Sanity: missing resource / action ----
		if resource == "" {
			return response.InternalError(c, "Authorization misconfigured", fiber.Map{
				"reason": "could not resolve resource from path",
				"path":   c.Path(),
			})
		}
		if action == "" {
			return response.InternalError(c, "Authorization misconfigured", fiber.Map{
				"reason": "could not resolve action from method",
				"path":   c.Path(),
				"method": c.Method(),
			})
		}

		// Store resolved domain for downstream consumers.
		c.Locals(authdomain.ContextKeyDomain, domain)

		// ---- 6. Casbin check ----
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
func isSelfServiceListAccounts(c fiber.Ctx) bool {
	if c.Method() != http.MethodGet {
		return false
	}
	path := strings.TrimSuffix(c.Path(), "/")
	return path == "/api/v1/accounts"
}

// isSelfServiceCreateAccount matches POST /api/v1/accounts/personal and
// POST /api/v1/accounts/institution.
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

func checkPermissionWithFallback(
	c fiber.Ctx,
	checker authdomain.PermissionChecker,
	userID, domain, resource, action string,
) (bool, error) {
	ctx := c.Context()

	switch action {
	case authdomain.ActionRead.String():
		return checkAny(ctx, checker, userID, domain, resource, []string{
			authdomain.ActionRead.String(),
			authdomain.ActionReadAll.String(),
			authdomain.ActionReadOwn.String(),
		})

	case authdomain.ActionUpdate.String():
		return checkAny(ctx, checker, userID, domain, resource, []string{
			authdomain.ActionUpdate.String(),
			authdomain.ActionUpdateAll.String(),
			authdomain.ActionUpdateOwn.String(),
		})

	case authdomain.ActionDelete.String():
		return checkAny(ctx, checker, userID, domain, resource, []string{
			authdomain.ActionDelete.String(),
			authdomain.ActionDeleteAll.String(),
			authdomain.ActionDeleteOwn.String(),
		})

	case authdomain.ActionPublishAll.String(), authdomain.ActionPublishOwn.String():
		return checkAny(ctx, checker, userID, domain, resource, []string{
			authdomain.ActionPublishAll.String(),
			authdomain.ActionPublishOwn.String(),
		})
	}

	return checker.HasPermission(ctx, userID, domain, resource, action)
}

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
// Precedence:
//
//  1. Locals override — set by an upstream middleware (e.g. a team guard
//     that loads the team and injects its account ID).
//
//  2. Query override (?account_id=X).
//
//  3. Slug routes → DomainDeferred (service authorizes).
//
//  4. /accounts/:accountId/* → account:<accountId>.
//
//  5. Platform routes (/admin, /platform, /system) → platform.
//
//  6. Token's account_id (fallback for routes with no explicit account
//     in the URL).
//
//  7. DomainDeferred as the final fallback.
//
// TEAM-SCOPED ROUTES: /teams/:teamId/* is not resolved here. Resolving a
// team ID to an account requires a DB read the middleware deliberately
// does not perform. Such routes hit step 7 and are authorized by the
// service layer.
func getDomainFromRequest(c fiber.Ctx) string {
	path := c.Path()

	// 1. Locals override
	if domain := c.Locals(authdomain.ContextKeyDomain); domain != nil {
		if s, ok := domain.(string); ok && s != "" {
			return s
		}
	}

	// 2. Query override
	if accountID := c.Query("account_id"); accountID != "" {
		return authdomain.AccountDomain(accountID)
	}

	// 3. Slug routes → defer
	if isSlugRoute(path) {
		return authdomain.DomainDeferred
	}

	// 4. /accounts/:accountId/* → account:<id>
	if id := c.Params("accountId"); id != "" {
		return authdomain.AccountDomain(id)
	}
	if id := c.Params("id"); id != "" && strings.Contains(path, "/accounts/") {
		return authdomain.AccountDomain(id)
	}

	// 5. Platform routes
	if isPlatformPath(path) {
		return authdomain.DomainPlatform
	}

	// 6. Token's account_id
	if domain := domainFromToken(c); domain != "" {
		return domain
	}

	// 7. Defer to service
	return authdomain.DomainDeferred
}

// domainFromToken reads account_id from Locals and returns the domain.
// Returns "" if no account context is present.
//
// NOTE: Team context from the token is IGNORED for authz purposes.
func domainFromToken(c fiber.Ctx) string {
	aidRaw := c.Locals(authdomain.ContextKeyAccountID)
	if aidRaw == nil {
		return ""
	}
	aid, ok := aidRaw.(string)
	if !ok || aid == "" {
		return ""
	}
	return authdomain.AccountDomain(aid)
}

// isSlugRoute reports whether the path contains a slug-based identifier.
//
// Examples:
//
//	GET /accounts/slug/acme-corp
//	GET /account-types/slug/premium
//	GET /users/slug/jane-doe/profile
func isSlugRoute(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, p := range parts {
		if p == "slug" && i+1 < len(parts) && parts[i+1] != "" {
			return true
		}
	}
	return false
}

// isPlatformPath reports whether the path is a platform-admin route.
func isPlatformPath(path string) bool {
	return strings.HasPrefix(path, "/api/v1/admin") ||
		strings.HasPrefix(path, "/api/v1/platform") ||
		strings.HasPrefix(path, "/api/v1/system")
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
	if i+1 >= len(segments) {
		return ""
	}
	switch segments[i+1] {
	case "members":
		return authdomain.ResourceMember.String()
	case "logo":
		return authdomain.ResourceAccount.String()
	}
	if i+2 < len(segments) {
		switch segments[i+2] {
		case "events", "event":
			return authdomain.ResourceEvent.String()
		case "profile":
			return authdomain.ResourceProfile.String()
		case "logo":
			return authdomain.ResourceAccount.String()
		}
	}
	return ""
}

func institutionSubResource(segments []string, i int) string {
	if i+1 < len(segments) && segments[i+1] == "logo" {
		return authdomain.ResourceInstitution.String()
	}
	if i+2 < len(segments) {
		switch segments[i+2] {
		case "events", "event":
			return authdomain.ResourceEvent.String()
		case "profile":
			return authdomain.ResourceProfile.String()
		case "logo":
			return authdomain.ResourceInstitution.String()
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

func getActionFromRequest(c fiber.Ctx) string {
	path := c.Path()
	method := c.Method()

	if method == http.MethodPost && strings.HasSuffix(path, "/leave") {
		return authdomain.ActionLeave.String()
	}

	if (method == http.MethodPut || method == http.MethodPatch) &&
		strings.HasSuffix(path, "/role") {
		return authdomain.ActionUpdate.String()
	}

	if method == http.MethodPost && strings.Contains(path, "/invitations") {
		return authdomain.ActionInvite.String()
	}

	if strings.Contains(path, "/avatar") || strings.Contains(path, "/logo") {
		switch method {
		case http.MethodPost, http.MethodPut:
			return authdomain.ActionUpdate.String()
		case http.MethodDelete:
			return authdomain.ActionDelete.String()
		}
	}

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