// internal/modules/auth/middleware.go

package authmiddleware

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates JWT from cookie or Authorization header
func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		var tokenString string

		// 1. Try to get token from cookie (preferred - more secure)
		tokenCookie := c.Cookies("access_token")
		if tokenCookie != "" {
			tokenString = tokenCookie
		} else {
			// 2. Fallback: Get token from Authorization header
			authHeader := c.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}
		}

		// 3. No token found
		if tokenString == "" {
			return response.Unauthorized(c, "Authentication required", fiber.Map{
				"reason": "no access token found in cookie or Authorization header",
			})
		}

		// Parse and validate token - with explicit algorithm validation
		token, err := jwt.Parse(tokenString,
			func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			},
			jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}),
		)

		if err != nil || !token.Valid {
			return response.Unauthorized(c, "Invalid or expired token", fiber.Map{
				"error": err.Error(),
			})
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return response.Unauthorized(c, "Invalid token claims", nil)
		}

		// Check token type
		tokenType, _ := claims["type"].(string)
		if tokenType != "access" {
			return response.Unauthorized(c, "Invalid token type", fiber.Map{
				"expected": "access",
				"got":      tokenType,
			})
		}

		// Get user ID from claims
		userID, ok := claims["user_id"].(string)
		if !ok {
			return response.Unauthorized(c, "User ID not found in token", nil)
		}

		// Get user role if present
		userRole, _ := claims["role"].(string)
		userEmail, _ := claims["email"].(string)
		userName, _ := claims["name"].(string)

		// Set user context using permissions package constants
		c.Locals(authorization.ContextKeyUserID, userID)
		c.Locals(authorization.ContextKeyUserRole, userRole)
		c.Locals(authorization.ContextKeyUserEmail, userEmail)
		c.Locals(authorization.ContextKeyUserName, userName)

		return c.Next()
	}
}

// OptionalAuthMiddleware validates JWT if present but doesn't require it
// Useful for endpoints that work for both authenticated and unauthenticated users
func OptionalAuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		var tokenString string

		// 1. Try to get token from cookie
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
			// No token, but that's okay for optional auth
			return c.Next()
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString,
			func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			},
			jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}),
		)

		if err != nil || !token.Valid {
			// Invalid token, but that's okay for optional auth
			return c.Next()
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Next()
		}

		tokenType, _ := claims["type"].(string)
		if tokenType != "access" {
			return c.Next()
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			return c.Next()
		}

		userRole, _ := claims["role"].(string)
		userEmail, _ := claims["email"].(string)
		userName, _ := claims["name"].(string)

		c.Locals(authorization.ContextKeyUserID, userID)
		c.Locals(authorization.ContextKeyUserRole, userRole)
		c.Locals(authorization.ContextKeyUserEmail, userEmail)
		c.Locals(authorization.ContextKeyUserName, userName)

		return c.Next()
	}
}