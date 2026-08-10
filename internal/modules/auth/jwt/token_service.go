package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
    config *config.Config
}

func NewTokenService(cfg *config.Config) domain.TokenService {
    return &TokenService{config: cfg}
}

func (s *TokenService) GenerateAccessToken(accountID string) (string, error) {
    claims := jwt.MapClaims{
        "account_id": accountID,
        "exp":        time.Now().Add(s.config.JWT.AccessExpiration).Unix(),
        "iat":        time.Now().Unix(),
        "type":       "access",
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.config.JWT.Secret))
}

func (s *TokenService) GenerateRefreshToken(accountID string) (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", fmt.Errorf("failed to generate refresh token: %w", err)
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}