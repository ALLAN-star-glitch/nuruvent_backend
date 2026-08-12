package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/email"
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

	// 4. Get account type
	accountType, err := s.repo.GetAccountTypeBySlug(userData["account_type"])
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

	// 8. Assign account_admin role in Casbin
	if err := s.permService.AssignAccountAdminRole(ctx, account.ID, account.ID); err != nil {
		log.Printf("Failed to assign permissions: %v", err)
	}

	// 9. Generate tokens
	accessToken, refreshToken, err := s.GenerateTokens(ctx, account)
	if err != nil {
		return nil, nil, err
	}

	// 10. ✅ Send Welcome Email via Queue
	accountTypeStr := userData["account_type"]
	if err := s.enqueueWelcomeEmail(account.Email, account.Name, accountTypeStr); err != nil {
		log.Printf("Failed to enqueue welcome email: %v", err)
		// Fallback: send synchronously
		if err := s.emailSvc.SendWelcome(account.Email, account.Name, accountTypeStr); err != nil {
			log.Printf("Failed to send welcome email: %v", err)
		}
	}

	// 11. Build response
	result := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    int64(s.config.JWT.AccessExpiration.Seconds()),
	}

	return account, result, nil
}

// enqueueWelcomeEmail enqueues a welcome email task
func (s *service) enqueueWelcomeEmail(to, name, accountType string) error {
	if s.queue == nil {
		return fmt.Errorf("queue client not available")
	}

	if accountType == "institution" {
		// Institution welcome requires institution name
		// You would need to fetch institution name from the account
		// For now, we'll use a generic institution name
		task := email.WelcomeInstitutionTask{
			To:              to,
			AdminName:       name,
			InstitutionName: "Your Institution", // Should fetch from DB
		}
		payload, err := task.Payload()
		if err != nil {
			return err
		}
		return s.queue.Enqueue(email.TypeWelcomeInstitution, payload)
	}

	// Individual welcome
	task := email.WelcomeIndividualTask{
		To:   to,
		Name: name,
	}
	payload, err := task.Payload()
	if err != nil {
		return err
	}
	return s.queue.Enqueue(email.TypeWelcomeIndividual, payload)
}