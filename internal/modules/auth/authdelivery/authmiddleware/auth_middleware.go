// internal/modules/auth/authdelivery/authmiddleware/middleware.go

package authmiddleware

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"

	"github.com/gofiber/fiber/v3"
)

// ============================================================
// TYPES
// ============================================================

// UserValidator reports whether a user ID corresponds to a live,
// active user. The middleware calls this after every successful JWT
// validation to catch ghost sessions (deleted users, DB resets,
// cross-environment tokens).
//
// It returns (true, nil) for a valid user, (false, nil) for a user
// that no longer exists, and (false, err) on infrastructure errors.
type UserValidator func(ctx context.Context, userID string) (bool, error)

// ============================================================
// AUTH MIDDLEWARE
// ============================================================

// AuthMiddleware validates JWT from cookie or Authorization header
// and verifies that the subject (user) still exists.
//
// If the token is structurally valid but references a user that no
// longer exists (deleted, DB reset, or issued by a different
// environment), the middleware clears the auth cookies and returns
// 401 so the frontend can force a re-login.
func AuthMiddleware(
	tokenService authdomain.TokenService,
	validateUser UserValidator,
	cfg *config.Config,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		tokenString := extractToken(c)
		if tokenString == "" {
			return response.Unauthorized(c, "Authentication required", nil)
		}

		tokenCtx, err := tokenService.ValidateToken(tokenString)
		if err != nil {
			// Structurally invalid or expired — no DB hit needed.
			return response.Unauthorized(c, "Invalid or expired token", nil)
		}

		// ---- Verify the user still exists in the current DB ----
		//
		// Prevents "ghost sessions": JWTs referencing users that were
		// deleted, soft-deleted, or never existed in this database.
		ok, err := validateUser(c.Context(), tokenCtx.UserID)
		if err != nil {
			log.Printf("⚠️  Auth: user validation error for subject %q: %v",
				tokenCtx.UserID, err)
			// Fail closed on infrastructure errors — treat as unauthenticated.
			clearAuthCookies(c, cfg)
			return response.Unauthorized(c, "Session expired. Please log in again.", nil)
		}
		if !ok {
			log.Printf("⚠️  Auth rejected: JWT subject %q no longer exists or is inactive",
				tokenCtx.UserID)
			clearAuthCookies(c, cfg)
			return response.Unauthorized(c, "Session expired. Please log in again.", nil)
		}

		// ---- Populate request context ----
		c.Locals(authdomain.ContextKeyUserID, tokenCtx.UserID)
		c.Locals(authdomain.ContextKeyUserRole, tokenCtx.Role)
		c.Locals(authdomain.ContextKeyUserEmail, tokenCtx.Email)
		c.Locals(authdomain.ContextKeyUserName, tokenCtx.DisplayName)

		c.Locals(authdomain.ContextKeyAccountID, tokenCtx.AccountID)
		c.Locals(authdomain.ContextKeyAccountType, tokenCtx.AccountTypeSlug)

		c.Locals(authdomain.ContextKeyTeamID, tokenCtx.TeamID)
		c.Locals(authdomain.ContextKeyTeamType, tokenCtx.TeamTypeSlug)

		return c.Next()
	}
}

// ============================================================
// OPTIONAL AUTH MIDDLEWARE
// ============================================================

// OptionalAuthMiddleware validates JWT if present but doesn't require
// it. A stale token is treated as "not authenticated" — cookies are
// cleared and the request continues anonymously.
func OptionalAuthMiddleware(
	tokenService authdomain.TokenService,
	validateUser UserValidator,
	cfg *config.Config,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		tokenString := extractToken(c)
		if tokenString == "" {
			return c.Next()
		}

		tokenCtx, err := tokenService.ValidateToken(tokenString)
		if err != nil {
			return c.Next()
		}

		ok, err := validateUser(c.Context(), tokenCtx.UserID)
		if err != nil || !ok {
			log.Printf("⚠️  Optional auth: JWT subject %q no longer valid, proceeding anonymous",
				tokenCtx.UserID)
			clearAuthCookies(c, cfg)
			return c.Next()
		}

		c.Locals(authdomain.ContextKeyUserID, tokenCtx.UserID)
		c.Locals(authdomain.ContextKeyUserRole, tokenCtx.Role)
		c.Locals(authdomain.ContextKeyUserEmail, tokenCtx.Email)
		c.Locals(authdomain.ContextKeyUserName, tokenCtx.DisplayName)

		c.Locals(authdomain.ContextKeyAccountID, tokenCtx.AccountID)
		c.Locals(authdomain.ContextKeyAccountType, tokenCtx.AccountTypeSlug)

		c.Locals(authdomain.ContextKeyTeamID, tokenCtx.TeamID)
		c.Locals(authdomain.ContextKeyTeamType, tokenCtx.TeamTypeSlug)

		return c.Next()
	}
}

// ============================================================
// HELPERS
// ============================================================

// extractToken pulls the JWT from the access_token cookie (preferred)
// or the Authorization header as a Bearer fallback.
func extractToken(c fiber.Ctx) string {
	if cookie := c.Cookies("access_token"); cookie != "" {
		return cookie
	}
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

// clearAuthCookies expires both auth cookies on the current response.
func clearAuthCookies(c fiber.Ctx, cfg *config.Config) {
	secure := cfg.Environment == "production"

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/auth/refresh",
	})
}

// ============================================================
// CONTEXT ACCESSORS
// ============================================================

// GetUserID extracts the user ID from the context.
func GetUserID(c fiber.Ctx) string {
	userID, ok := c.Locals(authdomain.ContextKeyUserID).(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUserRole extracts the user role from the context.
func GetUserRole(c fiber.Ctx) string {
	role, ok := c.Locals(authdomain.ContextKeyUserRole).(string)
	if !ok {
		return ""
	}
	return role
}

// GetUserEmail extracts the user email from the context.
func GetUserEmail(c fiber.Ctx) string {
	email, ok := c.Locals(authdomain.ContextKeyUserEmail).(string)
	if !ok {
		return ""
	}
	return email
}

// GetUserName extracts the user name from the context.
func GetUserName(c fiber.Ctx) string {
	name, ok := c.Locals(authdomain.ContextKeyUserName).(string)
	if !ok {
		return ""
	}
	return name
}

// GetAccountID extracts the account ID from the context.
//
// This is the only authorization-relevant scope carried by the token.
func GetAccountID(c fiber.Ctx) string {
	accountID, ok := c.Locals(authdomain.ContextKeyAccountID).(string)
	if !ok {
		return ""
	}
	return accountID
}

// GetAccountType extracts the account type from the context.
func GetAccountType(c fiber.Ctx) string {
	accountType, ok := c.Locals(authdomain.ContextKeyAccountType).(string)
	if !ok {
		return ""
	}
	return accountType
}

// GetAccountDomain returns the account domain for the current request.
//
// Format: "account:<account_id>". Returns "" if no account context exists.
func GetAccountDomain(c fiber.Ctx) string {
	accountID := GetAccountID(c)
	if accountID == "" {
		return ""
	}
	return authdomain.AccountDomain(accountID)
}

// GetTeamID extracts the team ID from the context.
//
// Team context is informational — it indicates which team the user is
// currently viewing. It is NOT used for authorization.
func GetTeamID(c fiber.Ctx) string {
	teamID, ok := c.Locals(authdomain.ContextKeyTeamID).(string)
	if !ok {
		return ""
	}
	return teamID
}

// GetTeamType extracts the team type from the context.
//
// Informational only. Not used for authorization.
func GetTeamType(c fiber.Ctx) string {
	teamType, ok := c.Locals(authdomain.ContextKeyTeamType).(string)
	if !ok {
		return ""
	}
	return teamType
}

// GetDomain extracts the resolved authz domain from the context.
//
// Set by the authorization middleware. Either "platform", "account:<id>",
// or "deferred" (slug routes where the service must authorize).
func GetDomain(c fiber.Ctx) string {
	domain, ok := c.Locals(authdomain.ContextKeyDomain).(string)
	if !ok {
		return ""
	}
	return domain
}

// GetUser extracts the current user context if available.
func GetUser(c fiber.Ctx) *authdomain.TokenContext {
	userID := GetUserID(c)
	if userID == "" {
		return nil
	}

	return &authdomain.TokenContext{
		UserID:       userID,
		Role:         GetUserRole(c),
		Email:        GetUserEmail(c),
		DisplayName:  GetUserName(c),
		AccountID:    GetAccountID(c),
		TeamID:       GetTeamID(c),
		TeamTypeSlug: GetTeamType(c),
		IsVerified:   false,
		IsActive:     true,
	}
}

// GetCurrentTeamID returns the team ID from the current context.
//
// Informational. Authorization uses the account domain.
func GetCurrentTeamID(c fiber.Ctx) string {
	return GetTeamID(c)
}

// IsPersonalTeamContext checks if the current request is in a personal team context.
//
// Informational only.
func IsPersonalTeamContext(c fiber.Ctx) bool {
	return GetTeamType(c) == "personal"
}

// IsInstitutionTeamContext checks if the current request is in an institution team context.
//
// Informational only.
func IsInstitutionTeamContext(c fiber.Ctx) bool {
	return GetTeamType(c) == "institution"
}

// HasTeamContext reports whether the current request carries a team context.
//
// Informational only.
func HasTeamContext(c fiber.Ctx) bool {
	return GetTeamID(c) != ""
}

// HasAccountContext reports whether the current request carries an account context.
func HasAccountContext(c fiber.Ctx) bool {
	return GetAccountID(c) != ""
}