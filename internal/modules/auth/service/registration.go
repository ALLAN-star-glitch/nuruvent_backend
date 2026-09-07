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


// Helper to extract strings from map[string]string safely
func getString(m map[string]string, key string) string {
    if val, ok := m[key]; ok {
        return val
    }
    return ""
}

// ============================================================
// REGISTRATION METHODS
// ============================================================

func (s *service) RegisterUser(ctx context.Context, req RegisterRequest) error {
	// 1. Check email uniqueness
	exists, err := s.repo.UserExistsByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if exists {
		return authdomain.ErrUserExists
	}

	// Check if user record already exists (e.g. inactive state)
	existingUser, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existingUser != nil {
		if existingUser.IsActive {
			otp := s.GenerateOTP()
			if err := s.StoreOTP(ctx, req.Email, otp, "registration"); err != nil {
				return fmt.Errorf("failed to store OTP: %w", err)
			}

			if err := s.notifSvc.SendOTP(ctx, authdomain.SendOTPRequest{
				To:      req.Email,
				Name:    existingUser.Name,
				OTP:     otp,
				Expires: "1 hour",
				Purpose: "registration",
			}); err != nil {
				log.Printf("[RegisterUser] Failed to send OTP notification: %v", err)
			}

			return fmt.Errorf("user with email '%s' already exists. A new OTP has been sent to your email", req.Email)
		}

		if err := s.repo.ReactivateUser(ctx, existingUser.ID); err != nil {
			return fmt.Errorf("failed to reactivate user: %w", err)
		}
	}

	// Check phone uniqueness
	exists, err = s.repo.UserExistsByPhone(ctx, req.Phone)
	if err != nil {
		return err
	}
	if exists {
		return authdomain.ErrInvalidPhone
	}

	// Generate & Store OTP
	otp := s.GenerateOTP()
	if err := s.StoreOTP(ctx, req.Email, otp, "registration"); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// Store user data in Redis for verification step
	userData := map[string]any{
		"email":             req.Email,
		"password":          req.Password,
		"name":              req.Name,
		"phone":             req.Phone,
		"account_type":      req.AccountType,
		"professional_type": req.ProfessionalType,
		"invite_token":      req.InviteToken, // Extracted if present
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

	// Dispatch OTP Email
	if err := s.notifSvc.SendOTP(ctx, authdomain.SendOTPRequest{
		To:      req.Email,
		Name:    req.Name,
		OTP:     otp,
		Expires: "1 hour",
		Purpose: "registration",
	}); err != nil {
		log.Printf("[RegisterUser] Failed to send OTP notification: %v", err)
	}

	return nil
}

// VerifyOTPAndCreateUser verifies OTP and creates user, accounts, and workspace teams.
func (s *service) VerifyOTPAndCreateUser(ctx context.Context, email, otp string) (*authdomain.User, map[string]any, error) {
	// 1. Verify OTP
	if err := s.VerifyOTP(ctx, email, otp, "registration"); err != nil {
		return nil, nil, err
	}

	// 2. Check existing active user
	existingUser, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		_ = s.DeleteOTP(ctx, email, "registration")
		_ = s.DeleteUserData(ctx, email)

		accessToken, refreshToken, err := s.GenerateTokens(ctx, existingUser)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
		}

		return existingUser, map[string]any{
			"user_id":       existingUser.ID,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"message":       "User already exists. Welcome back!",
		}, nil
	}

	// 3. Extract cached user payload from Redis
	userData, err := s.GetUserData(ctx, email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user data from cache: %w", err)
	}

	// Extract strongly typed string variables
	reqPassword := getString(userData, "password")
	reqName := getString(userData, "name")
	reqPhone := getString(userData, "phone")
	reqAccountType := getString(userData, "account_type")
	reqProfessionalType := getString(userData, "professional_type")
	reqInstitutionType := getString(userData, "institution_type")
	reqInstitutionName := getString(userData, "institution_name")
	reqInviteToken := getString(userData, "invite_token")

	// 4. Validate Account Type
	if reqAccountType == "" {
		reqAccountType = types.AccountTypePersonalName
	}
	if reqAccountType != types.AccountTypePersonalName && reqAccountType != types.AccountTypeInstitutionName {
		return nil, nil, fmt.Errorf("invalid account_type: %s", reqAccountType)
	}

	accountTypeID, err := s.getAccountTypeID(ctx, reqAccountType)
	if err != nil || accountTypeID == "" {
		return nil, nil, fmt.Errorf("failed to resolve account type ID: %w", err)
	}

	// 5. Validate Domain Enums
	var institutionTypeID *string
	if reqAccountType == types.AccountTypeInstitutionName {
		if reqInstitutionType == "" {
			return nil, nil, errors.New("institution_type is required for institution accounts")
		}
		institutionType, err := s.repo.GetInstitutionTypeByName(ctx, reqInstitutionType)
		if err != nil || institutionType == nil {
			return nil, nil, fmt.Errorf("invalid institution_type: %s", reqInstitutionType)
		}
		institutionTypeID = &institutionType.ID
	}

	var professionalTypeID *string
	if reqAccountType == types.AccountTypePersonalName && reqProfessionalType != "" {
		professionalType, err := s.repo.GetProfessionalTypeByName(ctx, reqProfessionalType)
		if err != nil || professionalType == nil {
			return nil, nil, fmt.Errorf("invalid professional_type: %s", reqProfessionalType)
		}
		professionalTypeID = &professionalType.ID
	}

	// 6. Password Hashing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 7. Sanitize values
	sanitizer := validation.Sanitize{}
	cleanName := sanitizer.DisplayName(reqName)
	cleanEmail := sanitizer.Identifier(email)
	cleanPhone := sanitizer.Identifier(reqPhone)
	accountSlug := sanitizer.GenerateSlugFromName(cleanName)

	user, err := authdomain.NewUser(cleanEmail, string(hashedPassword), cleanName, cleanPhone, accountTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to construct user entity: %w", err)
	}
	user.DisplayName = cleanName
	user.ProfessionalTypeID = professionalTypeID

	var account *authdomain.Account

	// ============================================================
	// ATOMIC TRANSACTION: User, Account, Account Member, Roles
	// ============================================================
	err = s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Create Core User
		if err := s.repo.CreateUser(txCtx, user); err != nil {
			return fmt.Errorf("failed to save user: %w", err)
		}

		// Create Core Account
		if reqAccountType == types.AccountTypePersonalName {
			account, err = authdomain.NewPersonalAccount(
				cleanName, cleanName, accountSlug, cleanEmail, cleanPhone, accountTypeID, user.ID,
			)
		} else {
			account, err = authdomain.NewInstitutionAccount(
				cleanName, cleanName, accountSlug, cleanEmail, cleanPhone, accountTypeID, *institutionTypeID, user.ID,
			)
		}
		if err != nil {
			return fmt.Errorf("failed to construct account: %w", err)
		}

		if err := s.repo.CreateAccount(txCtx, account); err != nil {
			return fmt.Errorf("failed to save account: %w", err)
		}

		// Create Account Membership
		accountMember, err := authdomain.NewAccountMember(
			account.ID, user.ID, authdomain.RoleAccountAdmin.String(), user.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to construct account member: %w", err)
		}

		if err := s.repo.CreateAccountMember(txCtx, accountMember); err != nil {
			return fmt.Errorf("failed to save account member: %w", err)
		}

		// Assign RBAC Policy
		accountDomain := authdomain.AccountDomain(account.ID)
		if err := s.roleManager.AssignRole(txCtx, accountDomain, user.ID, authdomain.RoleAccountAdmin.String()); err != nil {
			return fmt.Errorf("failed to assign account admin role: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("registration transaction failed: %w", err)
	}

	// ============================================================
	// WORKSPACE & INVITATION INTEGRATION
	// ============================================================

	// 1. Handle Invitation if token exists
	if reqInviteToken != "" {
		if _, err := s.teamSvc.AcceptInvitation(ctx, reqInviteToken, user.ID); err != nil {
			log.Printf("[VerifyOTPAndCreateUser] Warning: Failed to process invite token '%s': %v", reqInviteToken, err)
		}
	}

	// 2. Create Personal Team Workspace
	if err := s.createAndAddToPersonalTeam(ctx, user.ID, cleanName); err != nil {
		log.Printf("[VerifyOTPAndCreateUser] Warning: Failed to create personal team: %v", err)
	}

	// 3. Create Institution Team Workspace if applicable
	if reqAccountType == types.AccountTypeInstitutionName && account != nil {
		if err := s.createAndAddToInstitutionTeam(ctx, user.ID, account.ID, reqInstitutionName); err != nil {
			log.Printf("[VerifyOTPAndCreateUser] Warning: Failed to create institution team: %v", err)
		}
	}

	// ============================================================
	// TOKEN GENERATION & CLEANUP
	// ============================================================
	accessToken, refreshToken, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	_ = s.DeleteOTP(ctx, email, "registration")
	_ = s.DeleteUserData(ctx, email)

	// Build response object
	additionalData := map[string]any{
		"user_id":         user.ID,
		"access_token":    accessToken,
		"refresh_token":   refreshToken,
		"account_type":    reqAccountType,
		"account_type_id": accountTypeID,
		"account_id":      account.ID,
	}

	if account.IsInstitution() {
		additionalData["institution_type_id"] = account.InstitutionTypeID
		additionalData["institution_name"] = reqInstitutionName
		additionalData["institution_type"] = reqInstitutionType
	}

	if user.ProfessionalTypeID != nil {
		additionalData["professional_type_id"] = *user.ProfessionalTypeID
		additionalData["professional_type"] = reqProfessionalType
	}

	// Async Welcome Emails
	go func() {
		if err := s.sendWelcomeEmails(context.Background(), user, userData); err != nil {
			log.Printf("[VerifyOTPAndCreateUser] Failed to send welcome emails: %v", err)
		}
	}()

	return user, additionalData, nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

func (s *service) getAccountTypeID(ctx context.Context, accountType string) (string, error) {
	accountTypeObj, err := s.repo.GetAccountTypeByName(ctx, accountType)
	if err != nil || accountTypeObj == nil {
		return "", fmt.Errorf("failed to get account type: %w", err)
	}
	return accountTypeObj.ID, nil
}

func (s *service) createAndAddToPersonalTeam(ctx context.Context, userID, userName string) error {
	existingTeam, err := s.teamSvc.GetPersonalTeamByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check existing personal team: %w", err)
	}
	if existingTeam != nil {
		return nil
	}

	_, err = s.teamSvc.CreatePersonalTeam(ctx, userID, userName)
	if err != nil {
		return fmt.Errorf("failed to create personal team: %w", err)
	}

	return nil
}

func (s *service) createAndAddToInstitutionTeam(ctx context.Context, userID, accountID, institutionName string) error {
	if institutionName == "" {
		return errors.New("institution name is required")
	}

	sanitizer := validation.Sanitize{}
	displayName := sanitizer.DisplayName(institutionName)
	slug := sanitizer.GenerateSlugFromName(institutionName)

	_, err := s.teamSvc.CreateInstitutionTeam(ctx, accountID, institutionName, displayName, slug)
	if err != nil {
		return fmt.Errorf("failed to create institution team: %w", err)
	}

	return nil
}

func (s *service) sendWelcomeEmails(ctx context.Context, user *authdomain.User, userData map[string]string) error {
    if user == nil {
        return fmt.Errorf("user cannot be nil")
    }

    accountType := getString(userData, "account_type")
    adminEmail := s.config.NuruOnboardingNoticeEmails.AdminEmail

    switch accountType {
    case types.AccountTypePersonalName:
        _ = s.notifSvc.SendIndividualWelcome(ctx, authdomain.SendWelcomeRequest{
            To:   user.Email,
            Name: user.Name,
        })

        _ = s.notifSvc.SendNewPersonalAccountNotification(ctx, authdomain.SendNewPersonalAccountRegistrationRequest{
            To:                  adminEmail,
            NewAccountAdminName: user.Name,
        })

    case types.AccountTypeInstitutionName:
        instName := getString(userData, "institution_name")
        instEmail := getString(userData, "institution_email")

        _ = s.notifSvc.SendInstitutionWelcome(ctx, authdomain.SendInstitutionWelcomeRequest{
            To:               user.Email,
            AdminName:        user.Name,
            InstitutionName:  instName,
            InstitutionEmail: instEmail,
        })

        _ = s.notifSvc.SendNewInstitutionAccountNotification(ctx, authdomain.SendNewInstitutionAccountRegistrationRequest{
            To:                  adminEmail,
            NewAccountAdminName: user.Name,
            InstitutionName:     instName,
        })
    }

    return nil
}