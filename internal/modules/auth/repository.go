package auth

import (
	"errors"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ================================================
// USER OPERATIONS
// ================================================

// CreateUser creates a new user
func (r *Repository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// GetUserByEmail finds a user by email
func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByPhone finds a user by phone number
func (r *Repository) GetUserByPhone(phone string) (*models.User, error) {
	var user models.User
	err := r.db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID finds a user by ID (string)
func (r *Repository) GetUserByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByUUID finds a user by UUID
func (r *Repository) GetUserByUUID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmailOrPhone finds a user by email or phone
func (r *Repository) GetUserByEmailOrPhone(email, phone string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ? OR phone = ?", email, phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user
func (r *Repository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

// DeleteUser soft deletes a user
func (r *Repository) DeleteUser(id string) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

// ================================================
// USER CHECK OPERATIONS
// ================================================

// UserExistsByEmail checks if a user exists by email
func (r *Repository) UserExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UserExistsByPhone checks if a user exists by phone
func (r *Repository) UserExistsByPhone(phone string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("phone = ?", phone).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
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

// GetRefreshTokensByUserID finds all refresh tokens for a user
func (r *Repository) GetRefreshTokensByUserID(userID string) ([]models.RefreshToken, error) {
	var tokens []models.RefreshToken
	err := r.db.Where("user_id = ? AND revoked = ?", userID, false).
		Order("created_at DESC").
		Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// RevokeRefreshToken revokes a single refresh token
func (r *Repository) RevokeRefreshToken(token string) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("token = ?", token).
		Update("revoked", true).Error
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user
func (r *Repository) RevokeAllUserRefreshTokens(userID string) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("user_id = ?", userID).
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

// ================================================
// ADMIN OPERATIONS
// ================================================

// GetAllUsers gets all users with pagination
func (r *Repository) GetAllUsers(limit, offset int, search string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	db := r.db.Model(&models.User{})

	// Apply search filter if provided
	if search != "" {
		db = db.Where("email ILIKE ? OR name ILIKE ? OR phone ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := db.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

// GetUserStats gets user statistics
func (r *Repository) GetUserStats() (map[string]interface{}, error) {
	var totalUsers int64
	var totalAttendees int64
	var totalHosts int64

	// Total users
	if err := r.db.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	// Total attendees
	if err := r.db.Model(&models.User{}).Where("role = ?", "attendee").Count(&totalAttendees).Error; err != nil {
		return nil, err
	}

	// Total hosts
	if err := r.db.Model(&models.User{}).Where("role = ?", "host").Count(&totalHosts).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_users":     totalUsers,
		"total_attendees": totalAttendees,
		"total_hosts":     totalHosts,
	}

	return stats, nil
}