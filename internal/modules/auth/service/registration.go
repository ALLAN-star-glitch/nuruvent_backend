// internal/modules/auth/service/registration.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================
// REGISTRATION METHODS
// ============================================================

func (s *service) RegisterAccount(ctx context.Context, req RegisterRequest) error {
	// 1. Check email uniqueness
	exists, err := s.repo.AccountExistsByEmail(req.Email)
	if err != nil {
		return err
	}
	if exists {
		return authdomain.ErrAccountExists
	}

	// 2. Check phone uniqueness
	exists, err = s.repo.AccountExistsByPhone(req.Phone)
	if err != nil {
		return err
	}
	if exists {
		return authdomain.ErrInvalidPhone
	}

	// 3. Generate OTP
	otp := s.GenerateOTP()

	// 4. Store OTP with purpose
	if err := s.StoreOTP(ctx, req.Email, otp, "registration"); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// 5. Store user data
	userData := map[string]any{
		"email":        req.Email,
		"password":     req.Password,
		"name":         req.Name,
		"phone":        req.Phone,
		"account_type": req.AccountType,
	}

	if req.AccountType == "institution" {
		userData["institution_name"] = req.InstitutionName
		userData["institution_email"] = req.InstitutionEmail
		userData["institution_phone"] = req.InstitutionPhone
		userData["institution_type"] = req.InstitutionType
	}

	if err := s.StoreUserData(ctx, req.Email, userData); err != nil {
		return fmt.Errorf("failed to store user data: %w", err)
	}

	// 6. Send OTP via unified SendOTP
	if err := s.notifSvc.SendOTP(ctx, authdomain.SendOTPRequest{
		To:      req.Email,
		Name:    req.Name,
		OTP:     otp,
		Expires: "1 hour",
		Purpose: "registration",
		Meta:    nil,
	}); err != nil {
		log.Printf("Failed to send OTP notification: %v", err)
		// Don't fail the registration - OTP can be resent later
	}

	return nil
}

// VerifyOTPAndCreateAccount verifies OTP and creates the account
func (s *service) VerifyOTPAndCreateAccount(ctx context.Context, email, otp string) (*authdomain.Account, map[string]interface{}, error) {
	// 1. Verify OTP
	if err := s.VerifyOTP(ctx, email, otp, "registration"); err != nil {
		return nil, nil, err
	}

	// 2. Get user data
	userData, err := s.GetUserData(ctx, email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user data: %w", err)
	}

	// 3. Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData["password"]), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Get account type ID
	accountTypeID, err := s.getAccountTypeID(userData["account_type"])
	if err != nil {
		return nil, nil, err
	}

	// 5. Create account
	account, err := authdomain.NewAccount(
		userData["email"],
		string(hashedPassword),
		userData["name"],
		userData["phone"],
		accountTypeID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create account: %w", err)
	}

	if err := s.repo.CreateAccount(account); err != nil {
		return nil, nil, fmt.Errorf("failed to save account: %w", err)
	}

	// 6. Clean up
	if err := s.DeleteOTP(ctx, email, "registration"); err != nil {
		log.Printf("Failed to delete OTP: %v", err)
	}
	if err := s.DeleteUserData(ctx, email); err != nil {
		log.Printf("Failed to delete user data: %v", err)
	}

	// 7. Build additional data
	additionalData := map[string]interface{}{
		"account_id": account.ID,
	}

	if userData["account_type"] == "institution" {
		additionalData["institution_name"] = userData["institution_name"]
		additionalData["institution_email"] = userData["institution_email"]
		additionalData["institution_phone"] = userData["institution_phone"]
		additionalData["institution_type"] = userData["institution_type"]

		// Create institution record
		if err := s.createInstitution(ctx, account.ID, userData); err != nil {
			log.Printf("Failed to create institution: %v", err)
			// Don't fail the registration - institution can be created later
		}
	}

	// 8. Send welcome email
	if err := s.sendWelcomeEmail(ctx, account, userData); err != nil {
		log.Printf("Failed to send welcome email: %v", err)
	}

	return account, additionalData, nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

// getAccountTypeID returns the account type ID based on type string
func (s *service) getAccountTypeID(accountType string) (string, error) {
	switch accountType {
	case "personal":
		return "personal", nil
	case "institution":
		return "institution", nil
	default:
		return "", fmt.Errorf("invalid account type: %s", accountType)
	}
}

// createInstitution creates an institution record and links it to the account
func (s *service) createInstitution(ctx context.Context, accountID string, userData map[string]string) error {
	// 1. Get institution name
	institutionName := userData["institution_name"]
	if institutionName == "" {
		return errors.New("institution name is required")
	}

	// 2. Get institution email
	institutionEmail := userData["institution_email"]
	if institutionEmail == "" {
		return errors.New("institution email is required")
	}

	// 3. Get institution phone
	institutionPhone := userData["institution_phone"]
	if institutionPhone == "" {
		return errors.New("institution phone is required")
	}

	// 4. Get institution type
	institutionTypeSlug := userData["institution_type"]
	if institutionTypeSlug == "" {
		return errors.New("institution type is required")
	}

	// 5. Get institution type from database
	institutionType, err := s.repo.GetInstitutionTypeBySlug(institutionTypeSlug)
	if err != nil {
		return fmt.Errorf("failed to get institution type: %w", err)
	}
	if institutionType == nil {
		return authdomain.ErrInstitutionNotFound
	}

	// 6. Create institution
	institution, err := authdomain.NewInstitution(
		institutionName,
		institutionEmail,
		institutionPhone,
		institutionType.ID,
	)
	if err != nil {
		return err
	}

	// Set optional fields
	if description := userData["description"]; description != "" {
		institution.Description = description
	}
	if website := userData["website"]; website != "" {
		institution.Website = website
	}

	// 7. Save institution to database
	if err := s.repo.CreateInstitution(institution); err != nil {
		return fmt.Errorf("failed to create institution: %w", err)
	}

	// 8. Update account with institution ID
	if err := s.repo.UpdateAccountInstitutionID(accountID, institution.ID); err != nil {
		return fmt.Errorf("failed to update account with institution: %w", err)
	}

	// 9. Create team member (admin) using the existing account ID
	teamMember, err := authdomain.NewTeamMember(accountID, accountID, "admin")
	if err != nil {
		return err
	}

	if err := s.repo.CreateTeamMember(teamMember); err != nil {
		return fmt.Errorf("failed to create team member: %w", err)
	}

	return nil
}

// sendWelcomeEmail sends welcome email based on account type
func (s *service) sendWelcomeEmail(ctx context.Context, account *authdomain.Account, userData map[string]string) error {
	if userData["account_type"] == "institution" {
		return s.notifSvc.SendInstitutionWelcome(ctx, authdomain.SendInstitutionWelcomeRequest{
			To:              account.Email,
			AdminName:       account.Name,
			InstitutionName: userData["institution_name"],
		})
	}
	return s.notifSvc.SendIndividualWelcome(ctx, authdomain.SendWelcomeRequest{
		To:   account.Email,
		Name: account.Name,
	})
}