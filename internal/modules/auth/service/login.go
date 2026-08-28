// internal/modules/auth/service/login.go

package service

import (
	"context"
	"log"
	"time"

	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================
// LOGIN METHODS
// ============================================================

func (s *service) LoginUser(ctx context.Context, email, password, ipAddress, userAgent string) (*authdomain.User, string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", authdomain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", authdomain.ErrInvalidCredentials
	}

	if !user.IsActiveUser() {
		return nil, "", authdomain.ErrUserInactive
	}

	// Generate OTP
	otp := s.GenerateOTP()

	// Store 2FA OTP with purpose
	if err := s.StoreOTP(ctx, user.Email, otp, "two_factor"); err != nil {
		return nil, "", err
	}

	// Send 2FA OTP via unified SendOTP
	if err := s.notifSvc.SendOTP(ctx, authdomain.SendOTPRequest{
		To:      user.Email,
		Name:    user.Name,
		OTP:     otp,
		Expires: "5 minutes",
		Purpose: "two_factor",
		Meta: map[string]string{
			"ip_address": ipAddress,
			"user_agent": userAgent,
		},
	}); err != nil {
		log.Printf("Failed to send 2FA OTP: %v", err)
	}

	return user, otp, nil
}

func (s *service) VerifyTwoFactorAndLogin(ctx context.Context, email, otp, ipAddress, userAgent string) (*authdomain.User, string, string, error) {
	// Verify 2FA OTP using unified VerifyOTP
	if err := s.VerifyOTP(ctx, email, otp, "two_factor"); err != nil {
		return nil, "", "", err
	}

	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, "", "", err
	}
	if user == nil {
		return nil, "", "", authdomain.ErrUserNotFound
	}

	// Delete used 2FA OTP
	if err := s.DeleteOTP(ctx, email, "two_factor"); err != nil {
		log.Printf("Failed to delete 2FA OTP: %v", err)
	}

	// Generate tokens
	accessToken, refreshToken, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	// Send login notification
	now := time.Now().Format("January 2, 2006 at 3:04 PM")
	if err := s.notifSvc.SendLoginNotification(ctx, authdomain.SendLoginNotificationRequest{
		To:        user.Email,
		Name:      user.Name,
		Time:      now,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}); err != nil {
		log.Printf("Failed to send login notification: %v", err)
	}

	return user, accessToken, refreshToken, nil
}