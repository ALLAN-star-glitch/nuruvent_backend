// internal/modules/auth/authdomain/token_service.go

package authdomain

// OUTBOUND PORT - Domain defines what it needs
// The application layer depends on this interface
// Infrastructure (jwt package) implements this interface
type TokenService interface {
    GenerateAccessToken(ctx *TokenContext) (string, error)
    GenerateRefreshToken(userID string) (string, error)
    ValidateToken(tokenString string) (*TokenContext, error)
}