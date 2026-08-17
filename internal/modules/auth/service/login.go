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

func (s *service) LoginAccount(ctx context.Context, email, password, ipAddress, userAgent string) (*authdomain.Account, string, error) {
	account, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return nil, "", err
	}
	if account == nil {
		return nil, "", authdomain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
		return nil, "", authdomain.ErrInvalidCredentials
	}

	if !account.IsActiveAccount() {
		return nil, "", authdomain.ErrAccountInactive
	}

	// Generate OTP
	otp := s.GenerateOTP()

	// Store 2FA OTP with purpose
	if err := s.StoreOTP(ctx, account.Email, otp, "two_factor"); err != nil {
		return nil, "", err
	}

	// Send 2FA OTP via unified SendOTP
	if err := s.notifSvc.SendOTP(ctx, authdomain.SendOTPRequest{
		To:      account.Email,
		Name:    account.Name,
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

	return account, otp, nil
}

func (s *service) VerifyTwoFactorAndLogin(ctx context.Context, email, otp, ipAddress, userAgent string) (*authdomain.Account, string, string, error) {
	// Verify 2FA OTP using unified VerifyOTP
	if err := s.VerifyOTP(ctx, email, otp, "two_factor"); err != nil {
		return nil, "", "", err
	}

	account, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return nil, "", "", err
	}
	if account == nil {
		return nil, "", "", authdomain.ErrAccountNotFound
	}

	// Delete used 2FA OTP
	if err := s.DeleteOTP(ctx, email, "two_factor"); err != nil {
		log.Printf("Failed to delete 2FA OTP: %v", err)
	}

	// Generate tokens
	accessToken, refreshToken, err := s.GenerateTokens(ctx, account)
	if err != nil {
		return nil, "", "", err
	}

	// Send login notification
	now := time.Now().Format("January 2, 2006 at 3:04 PM")
	if err := s.notifSvc.SendLoginNotification(ctx, authdomain.SendLoginNotificationRequest{
		To:        account.Email,
		Name:      account.Name,
		Time:      now,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}); err != nil {
		log.Printf("Failed to send login notification: %v", err)
	}

	return account, accessToken, refreshToken, nil
}