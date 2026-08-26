// internal/modules/auth/service/registration.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
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

	if req.AccountType == types.AccountTypeInstitutionName {
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

	// ============================================================
	// ✅ Validate institution type BEFORE creating account
	// ============================================================
	if userData["account_type"] == types.AccountTypeInstitutionName {
		institutionTypeName := userData["institution_type"]
		if institutionTypeName == "" {
			return nil, nil, errors.New("institution_type is required for institution accounts")
		}
		
		// ✅ FIXED: Use GetInstitutionTypeByName (queries by NAME with underscores)
		institutionType, err := s.repo.GetInstitutionTypeByName(institutionTypeName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to validate institution type: %w", err)
		}
		if institutionType == nil {
			return nil, nil, fmt.Errorf("invalid institution_type: %s. Valid types: %v", 
				institutionTypeName, types.AllInstitutionTypeNames())
		}
		
		// ✅ Store the validated ID for later use
		validatedInstitutionTypeID := institutionType.ID
		_ = validatedInstitutionTypeID // Used later in createInstitution
	}

	// 3. Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData["password"]), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Get account type ID from database - ✅ Uses NAME
	accountTypeName := userData["account_type"]
	if accountTypeName == "" {
		accountTypeName = types.AccountTypePersonalName
	}
	
	accountTypeID, err := s.getAccountTypeID(accountTypeName)
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

	// ============================================================
	// ✅ Assign account_admin role to the account itself
	// ============================================================
	if err := s.permService.AssignAccountAdminRole(ctx, account.ID, account.ID); err != nil {
		log.Printf("⚠️ Failed to assign account_admin role: %v", err)
	}

	if err := s.permService.AddAccountPolicies(ctx, account.ID); err != nil {
		log.Printf("⚠️ Failed to add account policies: %v", err)
	}

	// ============================================================
	// ✅ Create institution (now we know the type is valid)
	// ============================================================
	if userData["account_type"] == types.AccountTypeInstitutionName {
		// ✅ FIXED: Get the validated institution type by NAME
		institutionTypeName := userData["institution_type"]
		institutionType, err := s.repo.GetInstitutionTypeByName(institutionTypeName)
		if err != nil {
			// This shouldn't happen since we validated above, but just in case
			return nil, nil, fmt.Errorf("failed to get institution type: %w", err)
		}
		if institutionType == nil {
			return nil, nil, fmt.Errorf("institution type not found: %s", institutionTypeName)
		}

		// ✅ Create institution record
		if err := s.createInstitution(ctx, account.ID, userData, institutionType.ID); err != nil {
			return nil, nil, fmt.Errorf("failed to create institution: %w", err)
		}
		
		log.Printf("✅ Institution created successfully for account: %s", account.ID)
	}

	// ============================================================
	// ✅ Generate tokens after account creation
	// ============================================================
	accessToken, refreshToken, err := s.GenerateTokens(ctx, account)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 6. Clean up
	if err := s.DeleteOTP(ctx, email, "registration"); err != nil {
		log.Printf("Failed to delete OTP: %v", err)
	}
	if err := s.DeleteUserData(ctx, email); err != nil {
		log.Printf("Failed to delete user data: %v", err)
	}

	// 7. Build additional data WITH TOKENS
	additionalData := map[string]interface{}{
		"account_id":     account.ID,
		"access_token":   accessToken,
		"refresh_token":  refreshToken,
	}

	if userData["account_type"] == types.AccountTypeInstitutionName {
		additionalData["institution_name"] = userData["institution_name"]
		additionalData["institution_email"] = userData["institution_email"]
		additionalData["institution_phone"] = userData["institution_phone"]
		additionalData["institution_type"] = userData["institution_type"]
		
		// ✅ Fetch the account again to get the updated institution_id
		updatedAccount, err := s.repo.GetAccountByID(account.ID)
		if err == nil && updatedAccount != nil {
			account = updatedAccount
			if account.InstitutionID != nil {
				additionalData["institution_id"] = *account.InstitutionID
			}
		}
	}

	// 8. Send welcome emails
	if err := s.sendWelcomeEmails(ctx, account, userData); err != nil {
		log.Printf("Failed to send welcome emails: %v", err)
	}

	return account, additionalData, nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

// getAccountTypeID queries the database by NAME (with underscores)
func (s *service) getAccountTypeID(accountType string) (string, error) {
	// Query the database for the account type by NAME (not slug)
	accountTypeObj, err := s.repo.GetAccountTypeByName(accountType)
	if err != nil {
		return "", fmt.Errorf("failed to get account type: %w", err)
	}
	if accountTypeObj == nil {
		return "", fmt.Errorf("account type not found: %s", accountType)
	}
	return accountTypeObj.ID, nil
}

// createInstitution creates an institution record and links it to the account
func (s *service) createInstitution(ctx context.Context, accountID string, userData map[string]string, institutionTypeID string) error {
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

	// 4. Create institution with the validated type ID
	institution, err := authdomain.NewInstitution(
		institutionName,
		institutionEmail,
		institutionPhone,
		institutionTypeID,
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

	// 5. Save institution to database
	if err := s.repo.CreateInstitution(institution); err != nil {
		return fmt.Errorf("failed to create institution: %w", err)
	}

	// 6. Update account with institution ID
	if err := s.repo.UpdateAccountInstitutionID(accountID, institution.ID); err != nil {
		return fmt.Errorf("failed to update account with institution: %w", err)
	}

	// 7. Create team member (admin) using the existing account ID
	teamMember, err := authdomain.NewTeamMember(accountID, accountID, "admin")
	if err != nil {
		return err
	}

	if err := s.repo.CreateTeamMember(teamMember); err != nil {
		return fmt.Errorf("failed to create team member: %w", err)
	}

	return nil
}

func (s *service) sendWelcomeEmails(ctx context.Context, account *authdomain.Account, userData map[string]string) error {
	// Send individual welcome to the admin (account email)
	if err := s.notifSvc.SendIndividualWelcome(ctx, authdomain.SendWelcomeRequest{
		To:   account.Email,
		Name: account.Name,
	}); err != nil {
		log.Printf("Failed to send individual welcome to admin: %v", err)
		// Consider whether to return this error or continue
		// return fmt.Errorf("failed to send individual welcome: %w", err)
	}

	// If institution account, send KYC welcome to institution email
	if userData["account_type"] == types.AccountTypeInstitutionName {
		// Send KYC welcome to institution
		if err := s.notifSvc.SendInstitutionKYCWelcome(ctx, authdomain.SendInstitutionKYCWelcomeRequest{
			To:              userData["institution_email"],
			AdminName:       account.Name,
			InstitutionName: userData["institution_name"],
			InstitutionType: userData["institution_type"],
		}); err != nil {
			log.Printf("Failed to send institution KYC welcome: %v", err)
			// Continue to send notification even if KYC welcome fails
		}

		// Send new account notification to admin
		if err := s.notifSvc.SendNewAccountNotification(ctx, authdomain.SendNewInstitutionAccountRegistrationRequest{
			TO:                  s.config.NuruOnboardingNoticeEmails.AdminEmail,
			NewAccountAdminName: account.Name,
			InstitutionName:     userData["institution_name"],
			InstitutionType:     userData["institution_type"],
		}); err != nil {
			log.Printf("Failed to send new account notification: %v", err)
			return fmt.Errorf("failed to send new account notification: %w", err)
		}
	}

	return nil
}


