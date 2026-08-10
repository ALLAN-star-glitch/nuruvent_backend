package service

import (
    "context"
    "fmt"
    "time"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/domain"
)

func (s *service) GenerateTokens(ctx context.Context, account *domain.Account) (string, string, error) {
    accessToken, err := s.tokenSvc.GenerateAccessToken(account.ID)
    if err != nil {
        return "", "", fmt.Errorf("failed to generate access token: %w", err)
    }

    refreshToken, err := s.tokenSvc.GenerateRefreshToken(account.ID)
    if err != nil {
        return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
    }

    // Store refresh token in database
    newToken, err := domain.NewRefreshToken(
        account.ID,
        refreshToken,
        "", // userAgent will be passed
        "", // ipAddress will be passed
        time.Now().Add(s.config.JWT.RefreshExpiration),
    )
    if err != nil {
        return "", "", err
    }

    if err := s.repo.CreateRefreshToken(newToken); err != nil {
        return "", "", fmt.Errorf("failed to store refresh token: %w", err)
    }

    return accessToken, refreshToken, nil
}

func (s *service) RefreshTokens(ctx context.Context, refreshToken, userAgent, ip string) (string, string, error) {
    // 1. Get token from database
    token, err := s.repo.GetRefreshTokenByToken(refreshToken)
    if err != nil {
        return "", "", err
    }
    if token == nil {
        return "", "", domain.ErrInvalidToken
    }

    // 2. Check if valid
    if token.IsExpired() || token.IsRevoked() {
        return "", "", domain.ErrInvalidToken
    }

    // 3. Revoke old token
    if err := s.repo.RevokeRefreshToken(refreshToken); err != nil {
        return "", "", err
    }

    // 4. Generate new tokens
    accessToken, err := s.tokenSvc.GenerateAccessToken(token.AccountID)
    if err != nil {
        return "", "", err
    }

    newRefreshToken, err := s.tokenSvc.GenerateRefreshToken(token.AccountID)
    if err != nil {
        return "", "", err
    }

    // 5. Store new refresh token
    newToken, err := domain.NewRefreshToken(
        token.AccountID,
        newRefreshToken,
        userAgent,
        ip,
        time.Now().Add(s.config.JWT.RefreshExpiration),
    )
    if err != nil {
        return "", "", err
    }

    if err := s.repo.CreateRefreshToken(newToken); err != nil {
        return "", "", fmt.Errorf("failed to store refresh token: %w", err)
    }

    return accessToken, newRefreshToken, nil
}

func (s *service) RevokeToken(ctx context.Context, refreshToken string) error {
    return s.repo.RevokeRefreshToken(refreshToken)
}