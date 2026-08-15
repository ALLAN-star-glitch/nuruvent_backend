// internal/modules/auth/service/login.go

package service

import (
	"context"
	"fmt"
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

	// Generate and store 2FA OTP
	otp := s.GenerateOTP()
	if err := s.StoreTwoFactorOTP(account.Email, otp); err != nil {
		return nil, "", fmt.Errorf("failed to store 2FA OTP: %w", err)
	}

	// Send 2FA OTP via notification service
	if err := s.sendTwoFactorOTP(account.Email, account.Name, otp, ipAddress, userAgent); err != nil {
		log.Printf("Failed to send 2FA OTP: %v", err)
	}

	return account, otp, nil
}

func (s *service) VerifyTwoFactorAndLogin(ctx context.Context, email, otp, ipAddress, userAgent string) (*authdomain.Account, string, string, error) {
	storedOTP, err := s.GetTwoFactorOTP(email)
	if err != nil {
		return nil, "", "", authdomain.ErrInvalidOTP
	}
	if otp != storedOTP {
		return nil, "", "", authdomain.ErrInvalidOTP
	}

	account, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return nil, "", "", err
	}
	if account == nil {
		return nil, "", "", authdomain.ErrAccountNotFound
	}

	// Delete used 2FA OTP
	if err := s.DeleteTwoFactorOTP(email); err != nil {
		log.Printf("Failed to delete 2FA OTP: %v", err)
	}

	// Generate tokens
	accessToken, refreshToken, err := s.GenerateTokens(ctx, account)
	if err != nil {
		return nil, "", "", err
	}

	// Send login notification
	now := time.Now().Format("January 2, 2006 at 3:04 PM")
	if err := s.sendLoginNotification(account.Email, account.Name, now, ipAddress, userAgent); err != nil {
		log.Printf("Failed to send login notification: %v", err)
	}

	return account, accessToken, refreshToken, nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

// sendTwoFactorOTP sends 2FA OTP via notification service
func (s *service) sendTwoFactorOTP(to, name, otp, ipAddress, userAgent string) error {
	return s.notifSvc.SendTwoFactorOTP(context.Background(), authdomain.SendTwoFactorRequest{
		To:        to,
		Name:      name,
		OTP:       otp,
		Expires:   "5 minutes",
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

// sendLoginNotification sends login notification via notification service
func (s *service) sendLoginNotification(to, name, timeStr, ipAddress, userAgent string) error {
	return s.notifSvc.SendLoginNotification(context.Background(), authdomain.SendLoginNotificationRequest{
		To:        to,
		Name:      name,
		Time:      timeStr,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}