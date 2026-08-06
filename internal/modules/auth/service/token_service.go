// internal/modules/auth/service/token_service.go

package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TokenService defines the token service interface
type TokenService interface {
	GenerateAccessToken(account *models.Account) (string, error)
	GenerateRefreshToken(accountID string, userAgent, ipAddress string) (string, error)
	ValidateRefreshToken(token string) (*models.RefreshToken, error)
	RevokeRefreshToken(token string) error
	RotateRefreshToken(oldToken string, userAgent, ipAddress string) (string, string, error)
}

// tokenService implements the TokenService interface
type tokenService struct {
	repo   domain.Repository // Changed from authRepo *Repository to domain.Repository
	config *config.Config
}

// NewTokenService creates a new token service
func NewTokenService(
	repo domain.Repository, // Changed from authRepo *Repository to domain.Repository
	cfg *config.Config,
) TokenService {
	return &tokenService{
		repo:   repo,
		config: cfg,
	}
}

// GenerateAccessToken creates a new access token (JWT)
func (s *tokenService) GenerateAccessToken(account *models.Account) (string, error) {
	// Get account type name
	accountTypeName := ""
	if account.AccountType.ID != uuid.Nil {
		accountTypeName = account.AccountType.Name
	}

	claims := jwt.MapClaims{
		"account_id":   account.ID.String(),
		"account_type": accountTypeName,
		"email":        account.Email,
		"name":         account.Name,
		"exp":          time.Now().Add(s.config.JWT.AccessExpiration).Unix(),
		"iat":          time.Now().Unix(),
		"type":         "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}

// GenerateRefreshToken creates a new refresh token
func (s *tokenService) GenerateRefreshToken(accountID string, userAgent, ipAddress string) (string, error) {
	// Parse string to uuid.UUID
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return "", fmt.Errorf("invalid account ID: %w", err)
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
		AccountID:  accountUUID,
		Token:      token,
		ExpiresAt:  time.Now().Add(s.config.JWT.RefreshExpiration),
		Revoked:    false,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
	}

	if err := s.repo.CreateRefreshToken(refreshToken); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return token, nil
}

// ValidateRefreshToken validates a refresh token
func (s *tokenService) ValidateRefreshToken(token string) (*models.RefreshToken, error) {
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
func (s *tokenService) RevokeRefreshToken(token string) error {
	return s.repo.RevokeRefreshToken(token)
}

// RotateRefreshToken invalidates the old token and generates a new one
func (s *tokenService) RotateRefreshToken(oldToken string, userAgent, ipAddress string) (string, string, error) {
	// Validate the old token
	refreshToken, err := s.ValidateRefreshToken(oldToken)
	if err != nil {
		return "", "", err
	}

	// Get the account using account ID
	account, err := s.repo.GetAccountByID(refreshToken.AccountID)
	if err != nil {
		return "", "", errors.New("account not found")
	}
	if account == nil {
		return "", "", errors.New("account not found")
	}

	// Revoke the old token
	if err := s.repo.RevokeRefreshToken(oldToken); err != nil {
		return "", "", fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Generate new tokens
	newRefreshToken, err := s.GenerateRefreshToken(account.ID.String(), userAgent, ipAddress)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	newAccessToken, err := s.GenerateAccessToken(account)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}