// internal/modules/auth/service/registration.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/validation"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ============================================================
// REGISTRATION METHODS
// ============================================================

func (s *service) RegisterUser(ctx context.Context, req RegisterRequest) error {
	// 1. Check email uniqueness
	exists, err := s.repo.UserExistsByEmail(req.Email)
	if err != nil {
		return err
	}
	if exists {
		return authdomain.ErrUserExists
	}

	// Check if user already exists
	existingUser, err := s.repo.GetUserByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existingUser != nil {
		// User already exists - check if they're active
		if existingUser.IsActive {
			// Generate and send new OTP
			otp := s.GenerateOTP()
			if err := s.StoreOTP(ctx, req.Email, otp, "registration"); err != nil {
				return fmt.Errorf("failed to store OTP: %w", err)
			}

			// Send OTP
			if err := s.notifSvc.SendOTP(ctx, authdomain.SendOTPRequest{
				To:      req.Email,
				Name:    existingUser.Name,
				OTP:     otp,
				Expires: "1 hour",
				Purpose: "registration",
				Meta:    nil,
			}); err != nil {
				log.Printf("Failed to send OTP notification: %v", err)
			}

			return fmt.Errorf("user with email '%s' already exists. A new OTP has been sent to your email", req.Email)
		}

		// User exists but is inactive - reactivate
		if err := s.repo.ReactivateUser(existingUser.ID); err != nil {
			return fmt.Errorf("failed to reactivate user: %w", err)
		}

		// Continue with OTP flow
	}

	// Check phone uniqueness
	exists, err = s.repo.UserExistsByPhone(req.Phone)
	if err != nil {
		return err
	}
	if exists {
		return authdomain.ErrInvalidPhone
	}

	// Generate OTP
	otp := s.GenerateOTP()

	// Store OTP with purpose
	if err := s.StoreOTP(ctx, req.Email, otp, "registration"); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// Store user data
	userData := map[string]any{
		"email":             req.Email,
		"password":          req.Password,
		"name":              req.Name,
		"phone":             req.Phone,
		"account_type":      req.AccountType,
		"professional_type": req.ProfessionalType,
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

	// Send OTP
	if err := s.notifSvc.SendOTP(ctx, authdomain.SendOTPRequest{
		To:      req.Email,
		Name:    req.Name,
		OTP:     otp,
		Expires: "1 hour",
		Purpose: "registration",
		Meta:    nil,
	}); err != nil {
		log.Printf("Failed to send OTP notification: %v", err)
	}

	return nil
}

// VerifyOTPAndCreateUser verifies OTP and creates a new user
func (s *service) VerifyOTPAndCreateUser(ctx context.Context, email, otp string) (*authdomain.User, map[string]interface{}, error) {
	// 1. Verify OTP
	if err := s.VerifyOTP(ctx, email, otp, "registration"); err != nil {
		return nil, nil, err
	}

	// Check if user already exists
	existingUser, err := s.repo.GetUserByEmail(email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		log.Printf("⚠️ User with email %s already exists", email)

		// Clean up Redis data
		_ = s.DeleteOTP(ctx, email, "registration")
		_ = s.DeleteUserData(ctx, email)

		// Generate new tokens for existing user
		accessToken, refreshToken, err := s.GenerateTokens(ctx, existingUser)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
		}

		additionalData := map[string]interface{}{
			"user_id":       existingUser.ID,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"message":       "User already exists. Welcome back!",
		}

		return existingUser, additionalData, nil
	}

	// 2. Get user data
	userData, err := s.GetUserData(ctx, email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user data: %w", err)
	}

	// ============================================================
	// VALIDATE ALL DATA FIRST - BEFORE CREATING USER
	// ============================================================

	// 3. Validate account type
	accountTypeName := userData["account_type"]
	if accountTypeName == "" {
		accountTypeName = types.AccountTypePersonalName
	}
	if accountTypeName != types.AccountTypePersonalName && accountTypeName != types.AccountTypeInstitutionName {
		return nil, nil, fmt.Errorf("invalid account_type: %s. Must be '%s' or '%s'",
			accountTypeName, types.AccountTypePersonalName, types.AccountTypeInstitutionName)
	}

	// 4. Get account type ID
	accountTypeID, err := s.getAccountTypeID(accountTypeName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get account type: %w", err)
	}
	if accountTypeID == "" {
		return nil, nil, fmt.Errorf("account type not found: %s", accountTypeName)
	}

	// 5. Validate professional type for personal accounts
	var professionalTypeID *string
	if accountTypeName == types.AccountTypePersonalName {
		professionalTypeName := userData["professional_type"]
		if professionalTypeName != "" {
			professionalType, err := s.repo.GetProfessionalTypeByName(professionalTypeName)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to validate professional type: %w", err)
			}
			if professionalType == nil {
				return nil, nil, fmt.Errorf("invalid professional_type: %s. Valid types: %v",
					professionalTypeName, types.AllProfessionalTypeNames())
			}
			professionalTypeID = &professionalType.ID
		}
	}

	// 6. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData["password"]), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 7. Create user
	displayName := userData["display_name"]
	if displayName == "" {
		displayName = userData["name"]
	}

	user, err := authdomain.NewUser(
		userData["email"],
		string(hashedPassword),
		userData["name"],
		userData["phone"],
		accountTypeID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	if displayName != "" {
		user.DisplayName = displayName
	}

	if professionalTypeID != nil {
		user.ProfessionalTypeID = professionalTypeID
	}

	// 8. Save user to database
	if err := s.repo.CreateUser(user); err != nil {
		return nil, nil, fmt.Errorf("failed to save user: %w", err)
	}

	log.Printf("✅ User created: %s", user.ID)

	// ============================================================
	// ✅ CREATE ACCOUNT
	// ============================================================

	sanitizer := validation.Sanitize{}
	accountSlug := sanitizer.GenerateSlugFromName(userData["name"])
	cleanEmail := sanitizer.Identifier(userData["email"])
	cleanPhone := sanitizer.Identifier(userData["phone"])
	cleanName := sanitizer.DisplayName(userData["name"])
	cleanDisplayName := displayName

	account, err := authdomain.NewAccount(
		cleanName,
		cleanDisplayName,
		accountSlug,
		cleanEmail,
		cleanPhone,
		accountTypeID,
		user.ID,
	)
	if err != nil {
		log.Printf("⚠️ Failed to create account for user: %v", err)
		// Don't fail the registration, just log the error
	} else {
		if err := s.repo.CreateAccount(ctx, account); err != nil {
			log.Printf("⚠️ Failed to save account: %v", err)
		} else {
			log.Printf("✅ Account created for user: %s (Account ID: %s)", user.ID, account.ID)

			// Add user as account_admin in account_members
			accountMember, err := authdomain.NewAccountMember(
				account.ID,
				user.ID,
				authdomain.RoleAccountAdmin.String(),
				user.ID,
			)
			if err != nil {
				log.Printf("⚠️ Failed to create account member: %v", err)
			} else {
				if err := s.repo.CreateAccountMember(ctx, accountMember); err != nil {
					log.Printf("⚠️ Failed to save account member: %v", err)
				} else {
					log.Printf("✅ User added as account_admin of account: %s", account.ID)
				}
			}

			// Assign Casbin role for account domain
			accountDomain := authdomain.AccountDomain(account.ID)
			if err := s.roleManager.AssignRole(ctx, accountDomain, user.ID, authdomain.RoleAccountAdmin.String()); err != nil {
				log.Printf("⚠️ Failed to assign account admin role: %v", err)
			}
		}
	}

	// ============================================================
	// ✅ GENERATE TOKENS
	// ============================================================
	accessToken, refreshToken, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 9. Clean up Redis data
	if err := s.DeleteOTP(ctx, email, "registration"); err != nil {
		log.Printf("Failed to delete OTP: %v", err)
	}
	if err := s.DeleteUserData(ctx, email); err != nil {
		log.Printf("Failed to delete user data: %v", err)
	}

	// 10. Build additional data
	additionalData := map[string]interface{}{
		"user_id":       user.ID,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	if account != nil {
		additionalData["account_id"] = account.ID
	}

	// Add professional type to response
	if user.ProfessionalTypeID != nil {
		additionalData["professional_type_id"] = *user.ProfessionalTypeID
	}
	if userData["professional_type"] != "" {
		additionalData["professional_type"] = userData["professional_type"]
	}

	// 11. Send welcome emails
	if err := s.sendWelcomeEmails(ctx, user, userData); err != nil {
		log.Printf("Failed to send welcome emails: %v", err)
	}

	return user, additionalData, nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

func (s *service) getAccountTypeID(accountType string) (string, error) {
	accountTypeObj, err := s.repo.GetAccountTypeByName(accountType)
	if err != nil {
		return "", fmt.Errorf("failed to get account type: %w", err)
	}
	if accountTypeObj == nil {
		return "", fmt.Errorf("account type not found: %s", accountType)
	}
	return accountTypeObj.ID, nil
}

// sendWelcomeEmails sends welcome emails based on account type
func (s *service) sendWelcomeEmails(ctx context.Context, user *authdomain.User, userData map[string]string) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	var errors []error
	accountType := userData["account_type"]

	switch accountType {
	case types.AccountTypePersonalName:
		// Send personal welcome
		if err := s.notifSvc.SendIndividualWelcome(ctx, authdomain.SendWelcomeRequest{
			To:   user.Email,
			Name: user.Name,
		}); err != nil {
			errMsg := fmt.Errorf("failed to send individual welcome: %w", err)
			log.Printf("[sendWelcomeEmails] %v", errMsg)
			errors = append(errors, errMsg)
		}

		// Send new personal account notification to admin
		if err := s.notifSvc.SendNewPersonalAccountNotification(ctx, authdomain.SendNewPersonalAccountRegistrationRequest{
			To:                  s.config.NuruOnboardingNoticeEmails.AdminEmail,
			NewAccountAdminName: user.Name,
		}); err != nil {
			errMsg := fmt.Errorf("failed to send new personal account notification: %w", err)
			log.Printf("[sendWelcomeEmails] %v", errMsg)
			errors = append(errors, errMsg)
		}

	case types.AccountTypeInstitutionName:
		// Sent to institution admin
		if err := s.notifSvc.SendInstitutionWelcome(ctx, authdomain.SendInstitutionWelcomeRequest{
			To:               user.Email,
			AdminName:        user.Name,
			InstitutionName:  userData["institution_name"],
			InstitutionEmail: userData["institution_email"],
		}); err != nil {
			errMsg := fmt.Errorf("failed to send institution welcome: %w", err)
			log.Printf("[sendWelcomeEmails] %v", errMsg)
			errors = append(errors, errMsg)
		}

		// Send notification to internal admin
		if err := s.notifSvc.SendNewInstitutionAccountNotification(ctx, authdomain.SendNewInstitutionAccountRegistrationRequest{
			To:                  s.config.NuruOnboardingNoticeEmails.AdminEmail,
			NewAccountAdminName: user.Name,
			InstitutionName:     userData["institution_name"],
		}); err != nil {
			errMsg := fmt.Errorf("failed to send new institution account notification: %w", err)
			log.Printf("[sendWelcomeEmails] %v", errMsg)
			errors = append(errors, errMsg)
		}

	default:
		log.Printf("[sendWelcomeEmails] Unknown account type: %s for user %s", accountType, user.Email)
		errors = append(errors, fmt.Errorf("unknown account type: %s", accountType))
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors while sending welcome emails", len(errors))
	}

	return nil
}