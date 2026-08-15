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

		// 4. Store user context using authdomain constants
		c.Locals(authdomain.ContextKeyUserID, tokenCtx.UserID)
		c.Locals(authdomain.ContextKeyUserRole, tokenCtx.Role)
		c.Locals(authdomain.ContextKeyUserEmail, tokenCtx.Email)
		c.Locals(authdomain.ContextKeyAccountID, tokenCtx.AccountID)

		// Store additional context
		if tokenCtx.InstitutionID != "" {
			c.Locals(authdomain.ContextKeyInstitutionID, tokenCtx.InstitutionID)
		}

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

		// Store user context using authdomain constants
		c.Locals(authdomain.ContextKeyUserID, tokenCtx.UserID)
		c.Locals(authdomain.ContextKeyUserRole, tokenCtx.Role)
		c.Locals(authdomain.ContextKeyUserEmail, tokenCtx.Email)
		c.Locals(authdomain.ContextKeyAccountID, tokenCtx.AccountID)

		if tokenCtx.InstitutionID != "" {
			c.Locals(authdomain.ContextKeyInstitutionID, tokenCtx.InstitutionID)
		}

		return c.Next()
	}
}

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

// GetAccountID extracts the account ID from the context
func GetAccountID(c fiber.Ctx) string {
	accountID, ok := c.Locals(authdomain.ContextKeyAccountID).(string)
	if !ok {
		return ""
	}
	return accountID
}

// GetInstitutionID extracts the institution ID from the context
func GetInstitutionID(c fiber.Ctx) string {
	institutionID, ok := c.Locals(authdomain.ContextKeyInstitutionID).(string)
	if !ok {
		return ""
	}
	return institutionID
}