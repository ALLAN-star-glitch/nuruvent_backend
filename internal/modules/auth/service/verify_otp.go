package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/domain"
	"golang.org/x/crypto/bcrypt"
)

func (s *service) VerifyOTPAndCreateAccount(ctx context.Context, email, otp string) (*domain.Account, map[string]interface{}, error) {
	// 1. Verify OTP
	storedOTP, err := s.otpSvc.GetOTP(email)
	if err != nil {
		return nil, nil, domain.ErrInvalidOTP
	}
	if otp != storedOTP {
		return nil, nil, domain.ErrInvalidOTP
	}

	// 2. Get user data
	userData, err := s.otpSvc.GetUserData(email)
	if err != nil {
		return nil, nil, errors.New("registration data not found")
	}

	// 3. Clean up Redis
	s.otpSvc.DeleteOTP(email)
	s.otpSvc.DeleteUserData(email)

	// 4. Get account type (always "personal")
	accountType, err := s.repo.GetAccountTypeBySlug("personal")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get account type: %w", err)
	}
	if accountType == nil {
		return nil, nil, domain.ErrAccountTypeNotFound
	}

	// 5. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData["password"]), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 6. Create account
	account, err := domain.NewAccount(
		userData["email"],
		string(hashedPassword),
		userData["name"],
		userData["phone"],
		accountType.ID,
	)
	if err != nil {
		return nil, nil, err
	}

	// 7. Save account
	if err := s.repo.CreateAccount(account); err != nil {
		return nil, nil, fmt.Errorf("failed to create account: %w", err)
	}

	// 8. ✅ Assign account_admin role in Casbin (not in account table)
	if err := s.permService.AssignAccountAdminRole(ctx, account.ID, account.ID); err != nil {
		log.Printf("Failed to assign permissions: %v", err)
		// Don't fail registration - permissions can be retried later
	}

	// 9. Generate tokens
	accessToken, refreshToken, err := s.GenerateTokens(ctx, account)
	if err != nil {
		return nil, nil, err
	}

	// 10. Send welcome email
	if err := s.emailSvc.SendWelcome(account.Email, account.Name); err != nil {
		log.Printf("Failed to send welcome email: %v", err)
	}

	// 11. Build response
	result := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    int64(s.config.JWT.AccessExpiration.Seconds()),
	}

	return account, result, nil
}