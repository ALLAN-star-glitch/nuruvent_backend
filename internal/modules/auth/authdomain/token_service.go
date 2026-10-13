// internal/modules/auth/authdomain/token_service.go

package authdomain

import "time"

// OUTBOUND PORT - Domain defines what it needs
// The application layer depends on this interface
// Infrastructure (jwt package) implements this interface
type TokenService interface {
	GenerateAccessToken(ctx *TokenContext) (string, error)

	// GenerateAccessTokenWithTTL mints an access token with an explicit
	// lifetime. Non-positive ttl falls back to the configured default.
	//
	// The JWT's exp claim reflects the ttl passed in. Callers that
	// report an expiry to their own callers (e.g. to set a cookie's
	// Expires attribute) must use the same ttl, or the cookie will
	// outlive the token and every request past the JWT's actual
	// lifetime will 401 with "Invalid or expired token".
	GenerateAccessTokenWithTTL(ctx *TokenContext, ttl time.Duration) (string, error)

	GenerateRefreshToken(userID string) (string, error)
	ValidateToken(tokenString string) (*TokenContext, error)
}