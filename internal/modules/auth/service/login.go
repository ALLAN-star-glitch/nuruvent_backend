package service

import (
	"context"
	"fmt"
	"log"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/domain"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *service) LoginAccount(ctx context.Context, email, password, ipAddress, userAgent string) (*domain.Account, string, error) {
    account, err := s.repo.GetAccountByEmail(email)
    if err != nil {
        return nil, "", err
    }
    if account == nil {
        return nil, "", domain.ErrInvalidCredentials
    }

    if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
        return nil, "", domain.ErrInvalidCredentials
    }

    if !account.IsActiveAccount() {
        return nil, "", domain.ErrAccountInactive
    }

    otp := s.otpSvc.GenerateOTP()
    if err := s.otpSvc.StoreTwoFactorOTP(account.Email, otp); err != nil {
        return nil, "", fmt.Errorf("failed to store 2FA OTP: %w", err)
    }

    if err := s.otpSvc.SendTwoFactorOTP(account.Email, account.Name, otp, ipAddress, userAgent); err != nil {
        log.Printf("Failed to send 2FA OTP: %v", err)
    }

    return account, otp, nil
}

func (s *service) VerifyTwoFactorAndLogin(ctx context.Context, email, otp, ipAddress, userAgent string) (*domain.Account, string, string, error) {
    storedOTP, err := s.otpSvc.GetTwoFactorOTP(email)
    if err != nil {
        return nil, "", "", domain.ErrInvalidOTP
    }
    if otp != storedOTP {
        return nil, "", "", domain.ErrInvalidOTP
    }

    account, err := s.repo.GetAccountByEmail(email)
    if err != nil {
        return nil, "", "", err
    }
    if account == nil {
        return nil, "", "", domain.ErrAccountNotFound
    }

    s.otpSvc.DeleteTwoFactorOTP(email)

    accessToken, refreshToken, err := s.GenerateTokens(ctx, account)
    if err != nil {
        return nil, "", "", err
    }

	now := time.Now().String()

    if err := s.emailSvc.SendLoginNotification(account.Email, account.Name, now, ipAddress, userAgent); err != nil {
        log.Printf("Failed to send login notification: %v", err)
    }

    return account, accessToken, refreshToken, nil
}