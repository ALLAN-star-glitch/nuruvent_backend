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

		// 4. Store user context
		c.Locals(authdomain.ContextKeyUserID, tokenCtx.UserID)
		c.Locals(authdomain.ContextKeyUserRole, tokenCtx.Role)
		c.Locals(authdomain.ContextKeyUserEmail, tokenCtx.Email)

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

		// Store user context
		c.Locals(authdomain.ContextKeyUserID, tokenCtx.UserID)
		c.Locals(authdomain.ContextKeyUserRole, tokenCtx.Role)
		c.Locals(authdomain.ContextKeyUserEmail, tokenCtx.Email)

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

// GetCurrentTeamID returns the team ID from the current context
func GetCurrentTeamID(c fiber.Ctx) string {
	domain := GetDomain(c)
	if authdomain.IsTeamDomain(domain) {
		return authdomain.ExtractTeamID(domain)
	}
	return ""
}

// IsPersonalTeamContext checks if the current request is in a personal team context
func IsPersonalTeamContext(c fiber.Ctx) bool {
	domain := GetDomain(c)
	return authdomain.IsTeamDomain(domain) && 
		domain == authdomain.PersonalTeamDomain(GetUserID(c))
}
