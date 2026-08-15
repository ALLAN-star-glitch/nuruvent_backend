// internal/modules/auth/jwt/token_service.go

package jwt

import (
	"fmt"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenService implements authdomain.TokenService using JWT
type TokenService struct {
	config *config.Config
}

// NewTokenService creates a new JWT token service
func NewTokenService(cfg *config.Config) authdomain.TokenService {
	return &TokenService{config: cfg}
}

// GenerateAccessToken generates a new access token from TokenContext
func (s *TokenService) GenerateAccessToken(ctx *authdomain.TokenContext) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":               ctx.UserID,
		"type":              "access",
		"iat":               now.Unix(),
		"exp":               now.Add(s.config.JWT.AccessExpiration).Unix(),
		"jti":               uuid.New().String(),
		"email":             ctx.Email,
		"role":              ctx.Role,
		"account_type_id":   ctx.AccountTypeID,
		"account_type_slug": ctx.AccountTypeSlug,
		"account_id":        ctx.AccountID,
	}

	// Add institution_id if present
	if ctx.InstitutionID != "" {
		claims["institution_id"] = ctx.InstitutionID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}

// GenerateRefreshToken generates a new refresh token
func (s *TokenService) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  userID,
		"type": "refresh",
		"iat":  now.Unix(),
		"exp":  now.Add(s.config.JWT.RefreshExpiration).Unix(),
		"jti":  uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}

// ValidateToken validates and parses a JWT token, returns TokenContext
func (s *TokenService) ValidateToken(tokenString string) (*authdomain.TokenContext, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWT.Secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Extract claims
	userID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string)
	accountTypeID, _ := claims["account_type_id"].(string)
	accountTypeSlug, _ := claims["account_type_slug"].(string)
	accountID, _ := claims["account_id"].(string)
	institutionID, _ := claims["institution_id"].(string)

	// Validate required fields
	if userID == "" {
		return nil, fmt.Errorf("user ID not found in token")
	}

	return &authdomain.TokenContext{
		UserID:          userID,
		Email:           email,
		Role:            role,
		AccountTypeID:   accountTypeID,
		AccountTypeSlug: accountTypeSlug,
		AccountID:       accountID,
		InstitutionID:   institutionID,
	}, nil
}