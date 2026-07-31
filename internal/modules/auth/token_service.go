// internal/modules/auth/token_service.go

package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
    "github.com/google/uuid"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type TokenService struct {
	repo   *Repository
	config *config.Config
}

func NewTokenService(repo *Repository, cfg *config.Config) *TokenService {
	return &TokenService{
		repo:   repo,
		config: cfg,
	}
}

// GenerateAccessToken creates a new access token (JWT)
func (s *TokenService) GenerateAccessToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"email":   user.Email,
		"name":    user.Name,
		"exp":     time.Now().Add(s.config.JWT.AccessExpiration).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}

// GenerateRefreshToken creates a new refresh token
func (s *TokenService) GenerateRefreshToken(userID string, userAgent, ipAddress string) (string, error) {
    // Parse string to uuid.UUID
    userUUID, err := uuid.Parse(userID)
    if err != nil {
        return "", fmt.Errorf("invalid user ID: %w", err)
    }

    // Generate 32 random bytes
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", fmt.Errorf("failed to generate refresh token: %w", err)
    }
    // Encode to Base64 URL-safe
    token := base64.URLEncoding.EncodeToString(bytes)

    // Store in database
    refreshToken := &models.RefreshToken{
        UserID:    userUUID,
        Token:     token,
        ExpiresAt: time.Now().Add(s.config.JWT.RefreshExpiration),
        Revoked:   false,
        UserAgent: userAgent,
        IPAddress: ipAddress,
    }

    if err := s.repo.CreateRefreshToken(refreshToken); err != nil {
        return "", fmt.Errorf("failed to store refresh token: %w", err)
    }

    return token, nil
}

// ValidateRefreshToken validates a refresh token
func (s *TokenService) ValidateRefreshToken(token string) (*models.RefreshToken, error) {
	refreshToken, err := s.repo.GetRefreshTokenByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid refresh token")
		}
		return nil, err
	}
	if refreshToken == nil {
		return nil, errors.New("invalid refresh token")
	}

	if refreshToken.Revoked {
		return nil, errors.New("refresh token has been revoked")
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, errors.New("refresh token has expired")
	}

	return refreshToken, nil
}

// RevokeRefreshToken revokes a single refresh token
func (s *TokenService) RevokeRefreshToken(token string) error {
	return s.repo.RevokeRefreshToken(token)
}

// RotateRefreshToken invalidates the old token and generates a new one
func (s *TokenService) RotateRefreshToken(oldToken string, userAgent, ipAddress string) (string, string, error) {
    // Validate the old token
    refreshToken, err := s.ValidateRefreshToken(oldToken)
    if err != nil {
        return "", "", err
    }

    // Get the user using string ID
    user, err := s.repo.GetUserByID(refreshToken.UserID.String())
    if err != nil {
        return "", "", errors.New("user not found")
    }
    if user == nil {
        return "", "", errors.New("user not found")
    }

    // Revoke the old token
    if err := s.repo.RevokeRefreshToken(oldToken); err != nil {
        return "", "", fmt.Errorf("failed to revoke old token: %w", err)
    }

    // Generate new tokens using string ID
    newRefreshToken, err := s.GenerateRefreshToken(user.ID.String(), userAgent, ipAddress)
    if err != nil {
        return "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
    }

    newAccessToken, err := s.GenerateAccessToken(user)
    if err != nil {
        return "", "", fmt.Errorf("failed to generate new access token: %w", err)
    }

    return newAccessToken, newRefreshToken, nil
}