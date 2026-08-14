package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/redis"
)

// ============================================================
// OTP METHODS (implements Service interface)
// ============================================================

func (s *service) GenerateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *service) StoreOTP(email, otp string) error {
	key := "otp:" + email
	return redis.Set(key, otp, 5*time.Minute)
}

func (s *service) GetOTP(email string) (string, error) {
	key := "otp:" + email
	otp, exists, err := redis.Get(key)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", domain.ErrOTPNotFound
	}
	return otp, nil
}

func (s *service) DeleteOTP(email string) error {
	key := "otp:" + email
	return redis.Delete(key)
}

func (s *service) SendOTPEmail(to, name, otp string) error {
	// Delegate to notification service
	return s.notifSvc.SendVerificationOTP(context.Background(), domain.SendOTPRequest{
		To:      to,
		Name:    name,
		OTP:     otp,
		Expires: "5 minutes",
		Purpose: "registration",
		Meta:    nil,
	})
}