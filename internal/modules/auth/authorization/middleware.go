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
// Resolution pipeline per request:
//
//	1. Extract caller identity (userID) — required.
//	2. Resolve the domain the caller is acting in.
//	3. Resolve resource + action from the path/method.
//	4. Short-circuit self-service routes.
//	5. Short-circuit deferred routes (slug lookups) — service authorizes.
//	6. Casbin check with fallback chain.
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
		// Casbin check is neither necessary nor meaningful. The resolver
		// would otherwise fall back to the token's team context and gate
		// the operation on the wrong domain.
		if isOwnProfileRequest(c) ||
			isOwnAvatarRequest(c) ||
			isInstitutionLogoRequest(c) ||
			isSelfServiceListAccounts(c) ||
			isSelfServiceCreateAccount(c) {
			log.Printf("✅ AUTHZ BYPASS: %s %s", c.Method(), c.Path())
			c.Locals(authdomain.ContextKeyDomain, domain)
			return c.Next()
		}

		// ---- 4. Deferred domain (slug-based lookups) ----
		//
		// The resolver returns DomainDeferred when the target cannot be
		// identified from the URL alone (e.g. /accounts/slug/acme-corp).
		// Passing the sentinel to Casbin would either always-deny or
		// always-allow. Instead we let the service layer perform the
		// lookup and then authorize against the resolved ID.
		if domain == authdomain.DomainDeferred {
			log.Printf("⏸  AUTHZ DEFERRED: %s %s (service must authorize)", c.Method(), c.Path())
			c.Locals(authdomain.ContextKeyDomain, domain)
			return c.Next()
		}

		// ---- 5. Sanity: missing resource / action ----
		//
		// If a route isn't mapped, silently passing "" to Casbin would
		// produce surprising results. Fail closed with a clear error.
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

	// Exact match for create, manage, leave, and any other action.
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
// Precedence (highest → lowest):
//
//  1. Explicit override via c.Locals(ContextKeyDomain)
//     — set by upstream middleware (e.g. after loading a resource)
//
//  2. Query param overrides (admin / cross-team tooling)
//       ?team_id=X&team_type=Y  → team:X (typed)
//       ?account_id=X           → account:X
//       ?scope=type:id          → type:X
//
//  3. Token team context (team_id + team_type from JWT)
//     — the default for authenticated requests
//
//  4. Slug routes → DomainDeferred (service authorizes)
//
//  5. Path-param resolution for RESOURCE-SCOPED routes
//     — institution → institution:team:<id>
//     — account     → account:<id>
//     — team        → team:<id> (typed)
//     — user        → personal:team:<id>
//
//  6. /teams/... path heuristic (UUID extraction)
//
//  7. Personal-scope routes (/me, /my, /profile, /avatar)
//
//  8. Platform routes (/admin, /platform, /system)
//
//  9. Fallback: caller's personal team
//
// NOTE: This function deliberately does NOT use AccountDomain(target) as
// both the domain and the resource context. The domain reflects the
// CALLER's scope; the resource is the thing being acted on. Mixing them
// collapses "which tenant am I acting in?" with "which object am I
// touching?" and produces self-referential Casbin checks that only pass
// when the caller is already inside the target — the opposite of what
// membership enforcement needs.
func getDomainFromRequest(c fiber.Ctx) string {
	path := c.Path()

	// ---- 1. Explicit override ----
	if domain := c.Locals(authdomain.ContextKeyDomain); domain != nil {
		if s, ok := domain.(string); ok && s != "" {
			return s
		}
	}

	// ---- 2. Query overrides ----
	if domain := domainFromQuery(c); domain != "" {
		return domain
	}

	// ---- 3. Token team context ----
	if domain := domainFromToken(c); domain != "" {
		return domain
	}

	// ---- 4. Slug routes → defer ----
	if isSlugRoute(path) {
		return authdomain.DomainDeferred
	}

	// ---- 5. Path-param resolution ----
	if domain := domainFromPathParams(c, path); domain != "" {
		return domain
	}

	// ---- 6. /teams/... heuristic ----
	if strings.Contains(path, "/teams/") {
		if domain := domainFromTeamsPath(c, path); domain != "" {
			return domain
		}
	}

	// ---- 7. Personal-scope routes ----
	if isPersonalScopePath(path) {
		if domain := personalDomainForCurrentUser(c); domain != "" {
			return domain
		}
	}

	// ---- 8. Platform routes ----
	if isPlatformPath(path) {
		return authdomain.DomainPlatform
	}

	// ---- 9. Fallback ----
	if domain := personalDomainForCurrentUser(c); domain != "" {
		return domain
	}

	return authdomain.DomainPlatform
}

// domainFromQuery honours ?team_id / ?team_type / ?account_id / ?scope.
func domainFromQuery(c fiber.Ctx) string {
	if teamID := c.Query("team_id"); teamID != "" {
		teamType := c.Query("team_type")
		if teamType == "" {
			teamType = authdomain.TeamTypePersonal
		}
		return authdomain.BuildTeamDomain(teamType, teamID)
	}
	if accountID := c.Query("account_id"); accountID != "" {
		return authdomain.AccountDomain(accountID)
	}
	if scope := c.Query("scope"); scope != "" {
		return domainFromScopeQuery(scope)
	}
	return ""
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
func domainFromScopeQuery(scope string) string {
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
//
// IMPORTANT: for account-scoped routes (e.g. /accounts/:id/members), this
// returns the ACCOUNT domain. The subsequent Casbin check therefore asks
// "does the caller have member:read *in account X's scope*?". For that to
// pass, the caller must have a role assignment in account X — which is
// exactly the membership model we want.
func domainFromPathParams(c fiber.Ctx, path string) string {
	// Explicit named params take precedence.
	if id := c.Params("institutionId"); id != "" {
		return authdomain.InstitutionTeamDomain(id)
	}
	if id := c.Params("teamId"); id != "" {
		return teamDomainWithType(c, id)
	}
	if id := c.Params("accountId"); id != "" {
		return authdomain.AccountDomain(id)
	}
	if id := c.Params("userId"); id != "" {
		return authdomain.PersonalTeamDomain(id)
	}

	// Generic :id — disambiguate by path family.
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
		if p != "teams" {
			continue
		}
		if i+1 < len(parts) && looksLikeUUID(parts[i+1]) {
			return parts[i+1]
		}
		if i+2 < len(parts) && looksLikeUUID(parts[i+2]) {
			return parts[i+2]
		}
	}
	return ""
}

// isSlugRoute reports whether the path contains a slug-based identifier
// that cannot be resolved to a domain without a DB lookup.
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

// isPersonalScopePath reports whether the path is explicitly personal-scoped.
//
// We use segment-based matching, NOT substring matching: a substring check
// for "/me" would falsely match any path containing "me" — e.g. a UUID
// segment, "/members", or "/media".
func isPersonalScopePath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for _, p := range parts {
		switch p {
		case "me", "my", "profile", "avatar":
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

// looksLikeUUID is a cheap UUID heuristic (dash + length).
func looksLikeUUID(s string) bool {
	return strings.Contains(s, "-") && len(s) >= 32
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
			// Fallback if the segment appears without an obvious owner.
			// account/institution sub-resolvers handle the common cases.
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

// accountSubResource inspects the segments immediately after "accounts".
//
// Path families handled:
//
//	/accounts/:id/members           → ResourceMember
//	/accounts/:id/members/:userId   → ResourceMember
//	/accounts/:id/logo              → ResourceAccount (explicit)
//	/accounts/:id/events            → ResourceEvent
//	/accounts/:id/profile           → ResourceProfile
func accountSubResource(segments []string, i int) string {
	if i+1 >= len(segments) {
		return ""
	}
	switch segments[i+1] {
	case "members":
		return authdomain.ResourceMember.String()
	case "logo":
		// Account logo is an account attribute, not a profile.
		// Returning ResourceAccount here avoids the accidental
		// dependency on loop ordering in the caller.
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
	if i+1 >= len(segments) && i+2 >= len(segments) {
		return ""
	}
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

// getActionFromRequest maps HTTP method + path suffix to a Casbin action.
//
// Verb overrides (in precedence order):
//
//	POST   .../leave             → leave
//	PUT    .../role              → update
//	POST   .../invitations       → invite
//	POST   .../avatar | /logo    → update
//	DELETE .../avatar | /logo    → delete
//
// Fallback: method-based mapping.
func getActionFromRequest(c fiber.Ctx) string {
	path := c.Path()
	method := c.Method()

	// ---- Verb-based overrides (path suffix) ----

	// POST /accounts/:id/leave → leave (membership self-removal).
	if method == http.MethodPost && strings.HasSuffix(path, "/leave") {
		return authdomain.ActionLeave.String()
	}

	// PUT/PATCH /accounts/:id/members/:userId/role → update.
	if (method == http.MethodPut || method == http.MethodPatch) &&
		strings.HasSuffix(path, "/role") {
		return authdomain.ActionUpdate.String()
	}

	// POST /invitations → invite.
	if method == http.MethodPost && strings.Contains(path, "/invitations") {
		return authdomain.ActionInvite.String()
	}

	// ---- Avatar / logo → update/delete, not create ----
	if strings.Contains(path, "/avatar") || strings.Contains(path, "/logo") {
		switch method {
		case http.MethodPost, http.MethodPut:
			return authdomain.ActionUpdate.String()
		case http.MethodDelete:
			return authdomain.ActionDelete.String()
		}
	}

	// ---- Method-based default ----
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