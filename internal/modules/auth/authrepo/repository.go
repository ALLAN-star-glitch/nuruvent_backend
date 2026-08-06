// internal/modules/auth/authrepo/repository.go

package authrepo

import (
	"errors"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ================================================
// REFRESH TOKEN OPERATIONS
// ================================================

// CreateRefreshToken creates a new refresh token
func (r *Repository) CreateRefreshToken(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

// GetRefreshTokenByToken finds a refresh token by its token string
func (r *Repository) GetRefreshTokenByToken(token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.db.Where("token = ?", token).First(&refreshToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &refreshToken, nil
}

// RevokeRefreshToken revokes a single refresh token
func (r *Repository) RevokeRefreshToken(token string) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("token = ?", token).
		Update("revoked", true).Error
}

// RevokeAllAccountRefreshTokens revokes all refresh tokens for an account
func (r *Repository) RevokeAllAccountRefreshTokens(accountID string) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("account_id = ?", accountID).
		Update("revoked", true).Error
}

// DeleteExpiredRefreshTokens deletes all expired refresh tokens
func (r *Repository) DeleteExpiredRefreshTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.RefreshToken{}).Error
}

// CleanupRevokedRefreshTokens deletes revoked tokens older than a certain time
func (r *Repository) CleanupRevokedRefreshTokens(olderThan time.Time) error {
	return r.db.Where("revoked = ? AND updated_at < ?", true, olderThan).
		Delete(&models.RefreshToken{}).Error
}