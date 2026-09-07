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
	log.Printf("[RegisterUser] Starting registration for email: %s, account_type: %s", req.Email, req.AccountType)

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
		"invite_token":      req.InviteToken,
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

	log.Printf("[RegisterUser] Registration initiated successfully for email: %s", req.Email)
	return nil
}


// VerifyOTPAndCreateUser verifies OTP and creates user, accounts, and workspace teams.
func (s *service) VerifyOTPAndCreateUser(ctx context.Context, email, otp string) (*authdomain.User, map[string]any, error) {
	log.Printf("[VerifyOTPAndCreateUser] Starting verification for email: %s", email)

	// 1. Verify OTP
	if err := s.VerifyOTP(ctx, email, otp, "registration"); err != nil {
		log.Printf("[VerifyOTPAndCreateUser] OTP verification failed: %v", err)
		return nil, nil, err
	}
	log.Printf("[VerifyOTPAndCreateUser] OTP verified successfully for email: %s", email)

	// 2. Check existing active user
	existingUser, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		log.Printf("[VerifyOTPAndCreateUser] User already exists: %s", existingUser.ID)
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
	log.Printf("[VerifyOTPAndCreateUser] Retrieved user data from cache")

	// Extract strongly typed string variables
	reqPassword := getString(userData, "password")
	reqName := getString(userData, "name")
	reqPhone := getString(userData, "phone")
	reqAccountType := getString(userData, "account_type")
	reqProfessionalType := getString(userData, "professional_type")
	reqInstitutionType := getString(userData, "institution_type")
	reqInstitutionName := getString(userData, "institution_name")
	reqInstitutionEmail := getString(userData, "institution_email")
	reqInstitutionPhone := getString(userData, "institution_phone")
	reqInviteToken := getString(userData, "invite_token")

	log.Printf("[VerifyOTPAndCreateUser] Account type: %s, Professional type: %s", reqAccountType, reqProfessionalType)

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
	log.Printf("[VerifyOTPAndCreateUser] Account type ID: %s", accountTypeID)

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
		log.Printf("[VerifyOTPAndCreateUser] Institution type ID: %s", *institutionTypeID)
	}

	var professionalTypeID *string
	if reqAccountType == types.AccountTypePersonalName && reqProfessionalType != "" {
		professionalType, err := s.repo.GetProfessionalTypeByName(ctx, reqProfessionalType)
		if err != nil || professionalType == nil {
			return nil, nil, fmt.Errorf("invalid professional_type: %s", reqProfessionalType)
		}
		professionalTypeID = &professionalType.ID
		log.Printf("[VerifyOTPAndCreateUser] Professional type ID: %s", *professionalTypeID)
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
	log.Printf("[VerifyOTPAndCreateUser] Starting database transaction...")
	err = s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Create Core User
		log.Printf("[VerifyOTPAndCreateUser] Creating user: %s", cleanEmail)
		if err := s.repo.CreateUser(txCtx, user); err != nil {
			log.Printf("[VerifyOTPAndCreateUser] Failed to create user: %v", err)
			return fmt.Errorf("failed to save user: %w", err)
		}
		log.Printf("[VerifyOTPAndCreateUser] User created: %s", user.ID)

		// Create Core Account
		if reqAccountType == types.AccountTypePersonalName {
			log.Printf("[VerifyOTPAndCreateUser] Creating personal account for user: %s", user.ID)
			account, err = authdomain.NewPersonalAccount(
				cleanName, cleanName, accountSlug, cleanEmail, cleanPhone, accountTypeID, user.ID,
			)
		} else {
			log.Printf("[VerifyOTPAndCreateUser] Creating institution account for user: %s", user.ID)
			
			// ✅ For institution accounts, use the institution name for display name and slug
			institutionDisplayName := sanitizer.DisplayName(reqInstitutionName)
			institutionSlug := sanitizer.GenerateSlugFromName(reqInstitutionName)
			
			account, err = authdomain.NewInstitutionAccount(
				reqInstitutionName,      // ✅ Institution name becomes account.Name
				institutionDisplayName,  // ✅ Display Name: "Tech Corp Ltd"
				institutionSlug,         // ✅ Slug: "tech-corp-ltd"
				reqInstitutionEmail,     // ✅ Institution email becomes account.Email
				reqInstitutionPhone,     // ✅ Institution phone becomes account.Phone
				accountTypeID,
				*institutionTypeID,
				user.ID,
			)
		}
		if err != nil {
			log.Printf("[VerifyOTPAndCreateUser] Failed to construct account: %v", err)
			return fmt.Errorf("failed to construct account: %w", err)
		}

		if err := s.repo.CreateAccount(txCtx, account); err != nil {
			log.Printf("[VerifyOTPAndCreateUser] Failed to save account: %v", err)
			return fmt.Errorf("failed to save account: %w", err)
		}
		log.Printf("[VerifyOTPAndCreateUser] Account created: %s", account.ID)

		// Create Account Membership
		log.Printf("[VerifyOTPAndCreateUser] Creating account member for user: %s", user.ID)
		accountMember, err := authdomain.NewAccountMember(
			account.ID, user.ID, authdomain.RoleAccountAdmin.String(), user.ID,
		)
		if err != nil {
			log.Printf("[VerifyOTPAndCreateUser] Failed to construct account member: %v", err)
			return fmt.Errorf("failed to construct account member: %w", err)
		}

		if err := s.repo.CreateAccountMember(txCtx, accountMember); err != nil {
			log.Printf("[VerifyOTPAndCreateUser] Failed to save account member: %v", err)
			return fmt.Errorf("failed to save account member: %w", err)
		}
		log.Printf("[VerifyOTPAndCreateUser] Account member created")

		// Assign RBAC Policy
		accountDomain := authdomain.AccountDomain(account.ID)
		log.Printf("[VerifyOTPAndCreateUser] Assigning account admin role for user: %s", user.ID)
		if err := s.roleManager.AssignRole(txCtx, accountDomain, user.ID, authdomain.RoleAccountAdmin.String()); err != nil {
			log.Printf("[VerifyOTPAndCreateUser] Failed to assign account admin role: %v", err)
			return fmt.Errorf("failed to assign account admin role: %w", err)
		}
		log.Printf("[VerifyOTPAndCreateUser] Account admin role assigned")

		return nil
	})

	if err != nil {
		log.Printf("[VerifyOTPAndCreateUser] Transaction failed: %v", err)
		return nil, nil, fmt.Errorf("registration transaction failed: %w", err)
	}
	log.Printf("[VerifyOTPAndCreateUser] Database transaction completed successfully")

	// ============================================================
	// WORKSPACE & INVITATION INTEGRATION
	// ============================================================

	// 1. Handle Invitation if token exists
	if reqInviteToken != "" {
		log.Printf("[VerifyOTPAndCreateUser] Processing invite token: %s", reqInviteToken)
		if _, err := s.teamSvc.AcceptInvitation(ctx, reqInviteToken, user.ID); err != nil {
			log.Printf("[VerifyOTPAndCreateUser] Warning: Failed to process invite token '%s': %v", reqInviteToken, err)
		} else {
			log.Printf("[VerifyOTPAndCreateUser] Invitation accepted successfully")
		}
	}

	// ============================================================
	// ✅ Create Personal Team ONLY for Personal Accounts
	// ============================================================
	if reqAccountType == types.AccountTypePersonalName {
		log.Printf("[VerifyOTPAndCreateUser] Creating personal team for user: %s (Personal Account)", user.ID)
		if err := s.createAndAddToPersonalTeam(ctx, user.ID, cleanName); err != nil {
			log.Printf("[VerifyOTPAndCreateUser] ⚠️ Warning: Failed to create personal team: %v", err)
		} else {
			log.Printf("[VerifyOTPAndCreateUser] ✅ Personal team created successfully for user: %s", user.ID)
		}
	} else {
		log.Printf("[VerifyOTPAndCreateUser] Skipping personal team creation for institution account: %s", reqAccountType)
	}

	// ============================================================
	// Create Institution Team Workspace ONLY for Institution Accounts
	// ============================================================
	if reqAccountType == types.AccountTypeInstitutionName && account != nil {
		log.Printf("[VerifyOTPAndCreateUser] Creating institution team for user: %s (Institution Account)", user.ID)
		if err := s.createAndAddToInstitutionTeam(ctx, user.ID, account.ID, reqInstitutionName); err != nil {
			log.Printf("[VerifyOTPAndCreateUser] ⚠️ Warning: Failed to create institution team: %v", err)
		} else {
			log.Printf("[VerifyOTPAndCreateUser] ✅ Institution team created successfully for user: %s", user.ID)
		}
	} else {
		log.Printf("[VerifyOTPAndCreateUser] Skipping institution team creation for personal account")
	}

	// ============================================================
	// TOKEN GENERATION & CLEANUP
	// ============================================================
	log.Printf("[VerifyOTPAndCreateUser] Generating tokens for user: %s", user.ID)
	accessToken, refreshToken, err := s.GenerateTokens(ctx, user)
	if err != nil {
		log.Printf("[VerifyOTPAndCreateUser] Failed to generate tokens: %v", err)
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	_ = s.DeleteOTP(ctx, email, "registration")
	_ = s.DeleteUserData(ctx, email)
	log.Printf("[VerifyOTPAndCreateUser] Cleaned up cache for email: %s", email)

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

	log.Printf("[VerifyOTPAndCreateUser] ✅ Registration completed successfully for user: %s, email: %s", user.ID, email)
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
	log.Printf("[createAndAddToPersonalTeam] Checking existing personal team for user: %s", userID)
	existingTeam, err := s.teamSvc.GetPersonalTeamByUserID(ctx, userID)
	if err != nil {
		log.Printf("[createAndAddToPersonalTeam] Failed to check existing personal team: %v", err)
		return fmt.Errorf("failed to check existing personal team: %w", err)
	}
	if existingTeam != nil {
		log.Printf("[createAndAddToPersonalTeam] Personal team already exists for user: %s, team_id: %s", userID, existingTeam.ID)
		return nil
	}

	log.Printf("[createAndAddToPersonalTeam] Creating personal team for user: %s", userID)
	team, err := s.teamSvc.CreatePersonalTeam(ctx, userID, userName)
	if err != nil {
		log.Printf("[createAndAddToPersonalTeam] Failed to create personal team: %v", err)
		return fmt.Errorf("failed to create personal team: %w", err)
	}
	log.Printf("[createAndAddToPersonalTeam] ✅ Personal team created: %s (Team ID: %s)", team.Name, team.ID)
	return nil
}

// createAndAddToInstitutionTeam creates an institution team and adds the user
func (s *service) createAndAddToInstitutionTeam(ctx context.Context, userID, accountID, institutionName string) error {
    if institutionName == "" {
        return errors.New("institution name is required")
    }

    sanitizer := validation.Sanitize{}
    displayName := sanitizer.DisplayName(institutionName)
    slug := sanitizer.GenerateSlugFromName(institutionName)

    log.Printf("[createAndAddToInstitutionTeam] Creating institution team for account: %s", accountID)
    
    // ✅ Pass the userID as the creator
    team, err := s.teamSvc.CreateInstitutionTeam(ctx, accountID, institutionName, displayName, slug, userID)
    if err != nil {
        log.Printf("[createAndAddToInstitutionTeam] Failed to create institution team: %v", err)
        return fmt.Errorf("failed to create institution team: %w", err)
    }
    log.Printf("[createAndAddToInstitutionTeam] ✅ Institution team created: %s (Team ID: %s)", team.Name, team.ID)

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