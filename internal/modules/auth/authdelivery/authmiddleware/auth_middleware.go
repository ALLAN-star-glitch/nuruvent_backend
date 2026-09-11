// internal/modules/auth/authdelivery/authmiddleware/middleware.go

package authmiddleware

import (
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"

	"github.com/gofiber/fiber/v3"
)

// AuthMiddleware validates JWT from cookie or Authorization header
func AuthMiddleware(tokenService authdomain.TokenService) fiber.Handler {
	return func(c fiber.Ctx) error {
		var tokenString string

		// 1. Try cookie
		tokenCookie := c.Cookies("access_token")
		if tokenCookie != "" {
			tokenString = tokenCookie
		} else {
			// 2. Try Authorization header
			authHeader := c.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}
		}

		if tokenString == "" {
			return response.Unauthorized(c, "Authentication required", nil)
		}

		// 3. Validate token
		tokenCtx, err := tokenService.ValidateToken(tokenString)
		if err != nil {
			return response.Unauthorized(c, "Invalid or expired token", nil)
		}

		// 4. Store all user context from token
		c.Locals(authdomain.ContextKeyUserID, tokenCtx.UserID)
		c.Locals(authdomain.ContextKeyUserRole, tokenCtx.Role)
		c.Locals(authdomain.ContextKeyUserEmail, tokenCtx.Email)
		c.Locals(authdomain.ContextKeyUserName, tokenCtx.DisplayName)
		
		// ✅ Store account and team context from token
		c.Locals(authdomain.ContextKeyAccountID, tokenCtx.AccountID)
		c.Locals(authdomain.ContextKeyAccountType, tokenCtx.AccountTypeSlug)
		c.Locals(authdomain.ContextKeyTeamID, tokenCtx.TeamID)
		c.Locals(authdomain.ContextKeyTeamType, tokenCtx.TeamTypeSlug)

		// Domain is set by authorization middleware based on the request path
		// Do NOT set it here

		return c.Next()
	}
}

// OptionalAuthMiddleware validates JWT if present but doesn't require it
func OptionalAuthMiddleware(tokenService authdomain.TokenService) fiber.Handler {
	return func(c fiber.Ctx) error {
		var tokenString string

		// Try cookie
		tokenCookie := c.Cookies("access_token")
		if tokenCookie != "" {
			tokenString = tokenCookie
		} else {
			// Try Authorization header
			authHeader := c.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}
		}

		if tokenString == "" {
			return c.Next()
		}

		// Validate token
		tokenCtx, err := tokenService.ValidateToken(tokenString)
		if err != nil {
			return c.Next()
		}

		// Store all user context from token
		c.Locals(authdomain.ContextKeyUserID, tokenCtx.UserID)
		c.Locals(authdomain.ContextKeyUserRole, tokenCtx.Role)
		c.Locals(authdomain.ContextKeyUserEmail, tokenCtx.Email)
		c.Locals(authdomain.ContextKeyUserName, tokenCtx.DisplayName)
		
		// ✅ Store account and team context from token
		c.Locals(authdomain.ContextKeyAccountID, tokenCtx.AccountID)
		c.Locals(authdomain.ContextKeyAccountType, tokenCtx.AccountTypeSlug)
		c.Locals(authdomain.ContextKeyTeamID, tokenCtx.TeamID)
		c.Locals(authdomain.ContextKeyTeamType, tokenCtx.TeamTypeSlug)

		// Domain is set by authorization middleware based on the request path
		// Do NOT set it here

		return c.Next()
	}
}

// ============================================================
// HELPER FUNCTIONS TO EXTRACT CONTEXT VALUES
// ============================================================

// GetUserID extracts the user ID from the context
func GetUserID(c fiber.Ctx) string {
	userID, ok := c.Locals(authdomain.ContextKeyUserID).(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUserRole extracts the user role from the context
func GetUserRole(c fiber.Ctx) string {
	role, ok := c.Locals(authdomain.ContextKeyUserRole).(string)
	if !ok {
		return ""
	}
	return role
}

// GetUserEmail extracts the user email from the context
func GetUserEmail(c fiber.Ctx) string {
	email, ok := c.Locals(authdomain.ContextKeyUserEmail).(string)
	if !ok {
		return ""
	}
	return email
}

// GetUserName extracts the user name from the context
func GetUserName(c fiber.Ctx) string {
	name, ok := c.Locals(authdomain.ContextKeyUserName).(string)
	if !ok {
		return ""
	}
	return name
}

// GetAccountID extracts the account ID from the context
func GetAccountID(c fiber.Ctx) string {
	accountID, ok := c.Locals(authdomain.ContextKeyAccountID).(string)
	if !ok {
		return ""
	}
	return accountID
}

// GetAccountType extracts the account type from the context
func GetAccountType(c fiber.Ctx) string {
	accountType, ok := c.Locals(authdomain.ContextKeyAccountType).(string)
	if !ok {
		return ""
	}
	return accountType
}

// GetTeamID extracts the team ID from the context
func GetTeamID(c fiber.Ctx) string {
	teamID, ok := c.Locals(authdomain.ContextKeyTeamID).(string)
	if !ok {
		return ""
	}
	return teamID
}

// GetTeamType extracts the team type from the context
func GetTeamType(c fiber.Ctx) string {
	teamType, ok := c.Locals(authdomain.ContextKeyTeamType).(string)
	if !ok {
		return ""
	}
	return teamType
}

// GetDomain extracts the domain from the context
func GetDomain(c fiber.Ctx) string {
	domain, ok := c.Locals(authdomain.ContextKeyDomain).(string)
	if !ok {
		return ""
	}
	return domain
}

// GetUser extracts the full User from the context (if available)
func GetUser(c fiber.Ctx) *authdomain.TokenContext {
	userID := GetUserID(c)
	if userID == "" {
		return nil
	}

	return &authdomain.TokenContext{
		UserID:      userID,
		Role:        GetUserRole(c),
		Email:       GetUserEmail(c),
		DisplayName: GetUserName(c),
		AccountID:   GetAccountID(c),
		TeamID:      GetTeamID(c),
		TeamTypeSlug: GetTeamType(c),
		IsVerified:  false,
		IsActive:    true,
	}
}

// GetUserTeamDomain returns the user's personal team domain
// Note: This returns the personal team domain based on the user ID
func GetUserTeamDomain(c fiber.Ctx) string {
	userID := GetUserID(c)
	if userID == "" {
		return ""
	}
	return authdomain.PersonalTeamDomain(userID)
}

// GetCurrentTeamDomain returns the team domain from the current context
func GetCurrentTeamDomain(c fiber.Ctx) string {
	teamID := GetTeamID(c)
	teamType := GetTeamType(c)
	if teamID == "" {
		return ""
	}
	if teamType == "institution" {
		return authdomain.InstitutionTeamDomain(teamID)
	}
	return authdomain.PersonalTeamDomain(teamID)
}


// GetCurrentTeamID returns the team ID from the current context
func GetCurrentTeamID(c fiber.Ctx) string {
	return GetTeamID(c)
}

// IsPersonalTeamContext checks if the current request is in a personal team context
func IsPersonalTeamContext(c fiber.Ctx) bool {
	teamType := GetTeamType(c)
	return teamType == "personal"
}

// IsInstitutionTeamContext checks if the current request is in an institution team context
func IsInstitutionTeamContext(c fiber.Ctx) bool {
	teamType := GetTeamType(c)
	return teamType == "institution"
}

// HasTeamContext checks if the current request has a team context
func HasTeamContext(c fiber.Ctx) bool {
	return GetTeamID(c) != ""
}

// HasAccountContext checks if the current request has an account context
func HasAccountContext(c fiber.Ctx) bool {
	return GetAccountID(c) != ""
}