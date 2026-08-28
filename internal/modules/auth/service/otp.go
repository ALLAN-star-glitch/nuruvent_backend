// internal/modules/auth/service/otp.go

package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// ============================================================
// OTP METHODS (implements Service interface)
// ============================================================

// GenerateOTP generates a 6-digit OTP
func (s *service) GenerateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// StoreOTP stores an OTP with purpose
func (s *service) StoreOTP(ctx context.Context, email, otp, purpose string) error {
	key := s.getOTPKey(email, purpose)
	expiry := s.getOTPExpiry(purpose)
	return s.redisClient.Set(ctx, key, otp, expiry)
}

// GetOTP retrieves an OTP by email and purpose
func (s *service) GetOTP(ctx context.Context, email, purpose string) (string, error) {
	key := s.getOTPKey(email, purpose)
	otp, exists, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", authdomain.ErrOTPNotFound
	}
	return otp, nil
}

// DeleteOTP deletes an OTP by email and purpose
func (s *service) DeleteOTP(ctx context.Context, email, purpose string) error {
	key := s.getOTPKey(email, purpose)
	return s.redisClient.Delete(ctx, key)
}

// VerifyOTP verifies an OTP for a specific purpose
func (s *service) VerifyOTP(ctx context.Context, email, otp, purpose string) error {
	storedOTP, err := s.GetOTP(ctx, email, purpose)
	if err != nil {
		return err
	}
	if storedOTP != otp {
		return authdomain.ErrInvalidOTP
	}
	return nil
}

// SendOTPEmail is a convenience method that generates, stores, and sends an OTP
func (s *service) SendOTPEmail(ctx context.Context, to, name, purpose string, meta map[string]string) error {
	// 1. Generate OTP
	otp := s.GenerateOTP()

	// 2. Store OTP
	if err := s.StoreOTP(ctx, to, otp, purpose); err != nil {
		return err
	}

	// 3. Send OTP via notification service
	return s.notifSvc.SendOTP(ctx, authdomain.SendOTPRequest{
		To:      to,
		Name:    name,
		OTP:     otp,
		Expires: s.getOTPExpiryString(purpose),
		Purpose: purpose,
		Meta:    meta,
	})
}

// ============================================================
// HELPER METHODS
// ============================================================

// getOTPKey returns the Redis key for OTP storage
func (s *service) getOTPKey(email, purpose string) string {
	return fmt.Sprintf("otp:%s:%s", purpose, email)
}

// getOTPExpiry returns the expiry duration based on purpose
func (s *service) getOTPExpiry(purpose string) time.Duration {
	switch purpose {
	case "registration", "email_change", "phone_change":
		return 1 * time.Hour
	case "two_factor":
		return 5 * time.Minute
	case "password_reset":
		return 15 * time.Minute
	default:
		return 5 * time.Minute
	}
}

// getOTPExpiryString returns a human-readable expiry string
func (s *service) getOTPExpiryString(purpose string) string {
	switch purpose {
	case "registration", "email_change", "phone_change":
		return "1 hour"
	case "two_factor":
		return "5 minutes"
	case "password_reset":
		return "15 minutes"
	default:
		return "5 minutes"
	}
}

// ============================================================
// USER DATA METHODS
// ============================================================

// StoreUserData stores temporary user data during registration
func (s *service) StoreUserData(ctx context.Context, email string, data map[string]interface{}) error {
	key := "user:data:" + email
	if err := s.redisClient.HSet(ctx, key, data); err != nil {
		return err
	}
	return s.redisClient.Expire(ctx, key, 1*time.Hour)
}

// GetUserData retrieves temporary user data during registration
func (s *service) GetUserData(ctx context.Context, email string) (map[string]string, error) {
	key := "user:data:" + email
	result, exists, err := s.redisClient.HGetAll(ctx, key)
	if err != nil {
		return nil, err
	}
	if !exists || len(result) == 0 {
		return nil, fmt.Errorf("user data not found")
	}
	return result, nil
}

// DeleteUserData deletes temporary user data during registration
func (s *service) DeleteUserData(ctx context.Context, email string) error {
	key := "user:data:" + email
	return s.redisClient.Delete(ctx, key)
}

// ============================================================
// PASSWORD RESET DATA METHODS
// ============================================================

// StoreResetData stores password reset data
func (s *service) StoreResetData(ctx context.Context, email, otp, newPassword string) error {
	key := "reset:" + email
	data := map[string]interface{}{
		"otp":          otp,
		"new_password": newPassword,
	}
	if err := s.redisClient.HSet(ctx, key, data); err != nil {
		return err
	}
	return s.redisClient.Expire(ctx, key, 15*time.Minute)
}

// GetResetData retrieves password reset data
func (s *service) GetResetData(ctx context.Context, email string) (map[string]string, error) {
	key := "reset:" + email
	result, exists, err := s.redisClient.HGetAll(ctx, key)
	if err != nil {
		return nil, err
	}
	if !exists || len(result) == 0 {
		return nil, fmt.Errorf("reset data not found")
	}
	return result, nil
}

// DeleteResetData deletes password reset data
func (s *service) DeleteResetData(ctx context.Context, email string) error {
	key := "reset:" + email
	return s.redisClient.Delete(ctx, key)
}

// ============================================================
// RESEND OTP
// ============================================================

// ResendOTP resends an OTP for any purpose
// Purpose can be: "registration", "two_factor", "password_reset", "email_change", "phone_change"
func (s *service) ResendOTP(ctx context.Context, email, name, purpose string) error {
	// 1. Validate purpose
	switch purpose {
	case "registration", "two_factor", "password_reset", "email_change", "phone_change":
		// Valid purpose
	default:
		return fmt.Errorf("invalid purpose: %s", purpose)
	}

	// 2. For certain purposes, try to get the user's name from the repository
	if name == "User" && (purpose == "two_factor" || purpose == "password_reset") {
		user, err := s.repo.GetUserByEmail(email)
		if err == nil && user != nil {
			name = user.Name
		}
		// Don't return error if user not found - just use default name
		// This prevents email enumeration attacks
	}

	// 3. Generate new OTP
	otp := s.GenerateOTP()

	// 4. Store OTP with purpose (overwrites existing)
	if err := s.StoreOTP(ctx, email, otp, purpose); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// 5. Send OTP via unified SendOTP
	if err := s.notifSvc.SendOTP(ctx, authdomain.SendOTPRequest{
		To:      email,
		Name:    name,
		OTP:     otp,
		Expires: s.getOTPExpiryString(purpose),
		Purpose: purpose,
		Meta:    nil,
	}); err != nil {
		return fmt.Errorf("failed to send OTP: %w", err)
	}

	return nil
}