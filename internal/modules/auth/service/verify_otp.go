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

	// 4. Get account type
	accountTypeSlug := userData["account_type"]
	if accountTypeSlug == "" {
		accountTypeSlug = "personal"
	}

	accountType, err := s.repo.GetAccountTypeBySlug(accountTypeSlug)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get account type: %w", err)
	}
	if accountType == nil {
		return nil, nil, domain.ErrAccountTypeNotFound
	}

	// 5. Hash password
	password := userData["password"]
	if password == "" {
		return nil, nil, errors.New("password not found")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 6. Create account
	accountEmail := userData["email"]
	if accountEmail == "" {
		return nil, nil, errors.New("email not found")
	}

	accountName := userData["name"]
	if accountName == "" {
		return nil, nil, errors.New("name not found")
	}

	accountPhone := userData["phone"]
	if accountPhone == "" {
		return nil, nil, errors.New("phone not found")
	}

	account, err := domain.NewAccount(
		accountEmail,
		string(hashedPassword),
		accountName,
		accountPhone,
		accountType.ID,
	)
	if err != nil {
		return nil, nil, err
	}

	// 7. Save account
	if err := s.repo.CreateAccount(account); err != nil {
		return nil, nil, fmt.Errorf("failed to create account: %w", err)
	}

	// 8. Handle institution creation
	var institution *domain.Institution
	if accountTypeSlug == "institution" {
		institution, err = s.createInstitution(ctx, userData, account.ID)
		if err != nil {
			return nil, nil, err
		}
		account.InstitutionID = &institution.ID

		if err := s.repo.UpdateAccount(account); err != nil {
			return nil, nil, fmt.Errorf("failed to update account with institution: %w", err)
		}
	}

	// 9. Assign account_admin role in Casbin
	if err := s.permService.AssignAccountAdminRole(ctx, account.ID, account.ID); err != nil {
		log.Printf("Failed to assign permissions: %v", err)
	}

	// 10. Generate tokens
	accessToken, refreshToken, err := s.GenerateTokens(ctx, account)
	if err != nil {
		return nil, nil, err
	}

	// 11. Send Welcome Email via Notification Service
	if err := s.sendWelcomeNotification(account, institution); err != nil {
		log.Printf("Failed to send welcome notification: %v", err)
	}

	// 12. Build response
	result := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    int64(s.config.JWT.AccessExpiration.Seconds()),
	}

	if institution != nil {
		result["institution"] = institution
	}

	return account, result, nil
}

// ============================================================
// INSTITUTION CREATION
// ============================================================

// createInstitution creates an institution from user data
func (s *service) createInstitution(_ context.Context, userData map[string]string, accountID string) (*domain.Institution, error) {
	// Get institution name
	institutionName := userData["institution_name"]
	if institutionName == "" {
		return nil, errors.New("institution name is required")
	}

	// Get institution email
	institutionEmail := userData["institution_email"]
	if institutionEmail == "" {
		return nil, errors.New("institution email is required")
	}

	// Get institution phone
	institutionPhone := userData["institution_phone"]
	if institutionPhone == "" {
		return nil, errors.New("institution phone is required")
	}

	// Get institution type
	institutionTypeSlug := userData["institution_type"]
	if institutionTypeSlug == "" {
		return nil, errors.New("institution type is required")
	}

	// Get institution type from database
	institutionType, err := s.repo.GetInstitutionTypeBySlug(institutionTypeSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get institution type: %w", err)
	}
	if institutionType == nil {
		return nil, domain.ErrInstitutionNotFound
	}

	// Create institution
	institution, err := domain.NewInstitution(
		institutionName,
		institutionEmail,
		institutionPhone,
		institutionType.ID,
	)
	if err != nil {
		return nil, err
	}

	// Set optional fields
	if description := userData["description"]; description != "" {
		institution.Description = description
	}
	if website := userData["website"]; website != "" {
		institution.Website = website
	}

	// Save institution to database
	if err := s.repo.CreateInstitution(institution); err != nil {
		return nil, fmt.Errorf("failed to create institution: %w", err)
	}

	// Create team member (admin) using the existing account ID
	teamMember, err := domain.NewTeamMember(accountID, accountID, "admin")
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateTeamMember(teamMember); err != nil {
		return nil, fmt.Errorf("failed to create team member: %w", err)
	}

	return institution, nil
}

// ============================================================
// WELCOME NOTIFICATION HELPERS
// ============================================================

// sendWelcomeNotification sends welcome notification via notification service
func (s *service) sendWelcomeNotification(account *domain.Account, institution *domain.Institution) error {
	// Check if it's an institution account
	if account.InstitutionID != nil && institution != nil {
		return s.notifSvc.SendInstitutionWelcome(context.Background(), domain.SendInstitutionWelcomeRequest{
			To:              account.Email,
			AdminName:       account.Name,
			InstitutionName: institution.Name,
		})
	}

	// Individual welcome
	return s.notifSvc.SendIndividualWelcome(context.Background(), domain.SendWelcomeRequest{
		To:   account.Email,
		Name: account.Name,
	})
}