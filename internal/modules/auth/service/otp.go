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

func (s *service) GenerateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *service) StoreOTP(email, otp string) error {
	ctx := context.Background()
	key := "otp:" + email
	return s.redisClient.Set(ctx, key, otp, 5*time.Minute)
}

func (s *service) GetOTP(email string) (string, error) {
	ctx := context.Background()
	key := "otp:" + email
	otp, exists, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", authdomain.ErrOTPNotFound
	}
	return otp, nil
}

func (s *service) DeleteOTP(email string) error {
	ctx := context.Background()
	key := "otp:" + email
	return s.redisClient.Delete(ctx, key)
}

func (s *service) SendOTPEmail(to, name, otp string) error {
	// Delegate to notification service
	return s.notifSvc.SendVerificationOTP(context.Background(), authdomain.SendOTPRequest{
		To:      to,
		Name:    name,
		OTP:     otp,
		Expires: "5 minutes",
		Purpose: "registration",
		Meta:    nil,
	})
}

// ============================================================
// TWO-FACTOR OTP METHODS
// ============================================================

func (s *service) StoreTwoFactorOTP(email, otp string) error {
	ctx := context.Background()
	key := "2fa:" + email
	return s.redisClient.Set(ctx, key, otp, 5*time.Minute)
}

func (s *service) GetTwoFactorOTP(email string) (string, error) {
	ctx := context.Background()
	key := "2fa:" + email
	otp, exists, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("2FA OTP not found or expired")
	}
	return otp, nil
}

func (s *service) DeleteTwoFactorOTP(email string) error {
	ctx := context.Background()
	key := "2fa:" + email
	return s.redisClient.Delete(ctx, key)
}

// ============================================================
// USER DATA METHODS
// ============================================================

func (s *service) StoreUserData(email string, data map[string]interface{}) error {
	ctx := context.Background()
	key := "user:data:" + email
	if err := s.redisClient.HSet(ctx, key, data); err != nil {
		return err
	}
	return s.redisClient.Expire(ctx, key, 5*time.Minute)
}

func (s *service) GetUserData(email string) (map[string]string, error) {
	ctx := context.Background()
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

func (s *service) DeleteUserData(email string) error {
	ctx := context.Background()
	key := "user:data:" + email
	return s.redisClient.Delete(ctx, key)
}

// ============================================================
// PASSWORD RESET DATA METHODS
// ============================================================

func (s *service) StoreResetData(email, otp, newPassword string) error {
	ctx := context.Background()
	key := "reset:" + email
	data := map[string]interface{}{
		"otp":          otp,
		"new_password": newPassword,
	}
	if err := s.redisClient.HSet(ctx, key, data); err != nil {
		return err
	}
	return s.redisClient.Expire(ctx, key, 5*time.Minute)
}

func (s *service) GetResetData(email string) (map[string]string, error) {
	ctx := context.Background()
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

func (s *service) DeleteResetData(email string) error {
	ctx := context.Background()
	key := "reset:" + email
	return s.redisClient.Delete(ctx, key)
}