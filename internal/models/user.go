package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email            string         `gorm:"uniqueIndex;not null" json:"email"`
	Phone            string         `gorm:"uniqueIndex" json:"phone,omitempty"`
	PasswordHash     string         `gorm:"column:password_hash;not null" json:"-"`
	Name             string         `gorm:"not null" json:"name"`
	Avatar           string         `gorm:"type:varchar(500)" json:"avatar,omitempty"`
	Role             string         `gorm:"not null;default:'attendee'" json:"role"`
	IsVerified       bool           `gorm:"default:false" json:"is_verified"`
	IsEmailVerified  bool           `gorm:"default:false" json:"is_email_verified"`
	VerifiedAt       *time.Time     `json:"verified_at,omitempty"`
	EmailVerifiedAt  *time.Time     `json:"email_verified_at,omitempty"`
	LastLoginAt      *time.Time     `json:"last_login_at,omitempty"`
	LastActiveAt     *time.Time     `json:"last_active_at,omitempty"`
	TwoFactorEnabled bool           `gorm:"default:false" json:"two_factor_enabled"`
	BusinessID       *uuid.UUID     `gorm:"type:uuid;index" json:"business_id,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Business *Business `gorm:"foreignKey:BusinessID" json:"business,omitempty"`
}

// HashPassword hashes the user's password
func (u *User) HashPassword(password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashed)
	return nil
}

// ComparePassword compares a plaintext password with the hashed password
func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

// TableName returns the table name
func (User) TableName() string {
	return "users"
}