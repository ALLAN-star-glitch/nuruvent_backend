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

	// 4. Get account type - ✅ No type assertion needed (already string)
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

	// 5. Hash password - ✅ No type assertion needed
	password := userData["password"]
	if password == "" {
		return nil, nil, errors.New("password not found")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 6. Create account - ✅ No type assertion needed
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

	// 7. Handle institution creation if applicable
	var institution *domain.Institution
	if accountTypeSlug == "institution" {
		institution, err = s.createInstitution(ctx, userData)
		if err != nil {
			return nil, nil, err
		}
		account.InstitutionID = &institution.ID
	}

	// 8. Save account
	if err := s.repo.CreateAccount(account); err != nil {
		return nil, nil, fmt.Errorf("failed to create account: %w", err)
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

	// 11. Send Welcome Email via Queue
	if err := s.enqueueWelcomeEmail(account, institution); err != nil {
		log.Printf("Failed to enqueue welcome email: %v", err)
		// Fallback: send synchronously
		if err := s.sendWelcomeEmailSync(account, institution); err != nil {
			log.Printf("Failed to send welcome email: %v", err)
		}
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
func (s *service) createInstitution(_ context.Context, userData map[string]string) (*domain.Institution, error) {
	// ✅ No type assertions needed - userData is map[string]string
	institutionName := userData["institution_name"]
	if institutionName == "" {
		return nil, errors.New("institution name is required")
	}

	institutionEmail := userData["institution_email"]
	if institutionEmail == "" {
		return nil, errors.New("institution email is required")
	}

	institutionPhone := userData["institution_phone"]
	if institutionPhone == "" {
		return nil, errors.New("institution phone is required")
	}

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

	// Create team member (admin)
	teamMember, err := domain.NewTeamMember(institution.ID, institution.ID, "admin")
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateTeamMember(teamMember); err != nil {
		return nil, fmt.Errorf("failed to create team member: %w", err)
	}

	return institution, nil
}

// ============================================================
// WELCOME EMAIL HELPERS
// ============================================================

// enqueueWelcomeEmail enqueues a welcome email task
func (s *service) enqueueWelcomeEmail(account *domain.Account, institution *domain.Institution) error {
	if s.queue == nil {
		return fmt.Errorf("queue client not available")
	}

	// Check if it's an institution account
	if account.InstitutionID != nil && institution != nil {
		task := email.WelcomeInstitutionTask{
			To:              account.Email,
			AdminName:       account.Name,
			InstitutionName: institution.Name,
		}
		payload, err := task.Payload()
		if err != nil {
			return err
		}
		return s.queue.Enqueue(email.TypeWelcomeInstitution, payload)
	}

	// Individual welcome
	task := email.WelcomeIndividualTask{
		To:   account.Email,
		Name: account.Name,
	}
	payload, err := task.Payload()
	if err != nil {
		return err
	}
	return s.queue.Enqueue(email.TypeWelcomeIndividual, payload)
}

// sendWelcomeEmailSync sends welcome email synchronously (fallback)
func (s *service) sendWelcomeEmailSync(account *domain.Account, institution *domain.Institution) error {
	if account.InstitutionID != nil && institution != nil {
		return s.emailSvc.SendInstitutionWelcome(account.Email, account.Name, institution.Name)
	}
	return s.emailSvc.SendIndividualWelcome(account.Email, account.Name)
}