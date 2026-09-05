

package service

import "context"

// ============================================================
// TOKEN SERVICE INTERFACE
// ============================================================

// TokenService defines the interface for token generation operations
// used by the profile module
type TokenService interface {
    // GenerateTokens generates both access and refresh tokens for a user
    GenerateTokens(ctx context.Context, userID string) (accessToken, refreshToken string, err error)
    
    // GenerateAccessToken generates only an access token for a user
    GenerateAccessToken(ctx context.Context, userID string) (string, error)
}