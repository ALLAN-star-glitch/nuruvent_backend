package domain

// ============================================================
// OUTBOUND PORT: TokenService
// ============================================================

type TokenService interface {
    GenerateAccessToken(accountID string) (string, error)
    GenerateRefreshToken(accountID string) (string, error)
}