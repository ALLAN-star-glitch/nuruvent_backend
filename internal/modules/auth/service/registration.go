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

func (s *service) RegisterUser(ctx context.Context, req RegisterRequest) error {
	// 1. Check email uniqueness
	exists, err := s.repo.UserExistsByEmail(req.Email)
	if err != nil {
		return err
	}
	if exists {
		return authdomain.ErrUserExists
	}

	// 2. Check phone uniqueness
	exists, err = s.repo.UserExistsByPhone(req.Phone)
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
		"email":              req.Email,
		"password":           req.Password,
		"name":               req.Name,
		"phone":              req.Phone,
		"account_type":       req.AccountType,
		"professional_type":  req.ProfessionalType,
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

	// 6. Send OTP
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

	// 2. Get user data
	userData, err := s.GetUserData(ctx, email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user data: %w", err)
	}

	// ============================================================
	// ✅ VALIDATE ALL DATA FIRST - BEFORE CREATING USER
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

	// 4. Get account type ID (validate it exists)
	accountTypeID, err := s.getAccountTypeID(accountTypeName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get account type: %w", err)
	}
	if accountTypeID == "" {
		return nil, nil, fmt.Errorf("account type not found: %s", accountTypeName)
	}

	// 5. Validate institution type for institution accounts
	var institutionTypeID string
	if accountTypeName == types.AccountTypeInstitutionName {
		institutionTypeName := userData["institution_type"]
		if institutionTypeName == "" {
			return nil, nil, errors.New("institution_type is required for institution accounts")
		}

		institutionType, err := s.repo.GetInstitutionTypeByName(institutionTypeName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to validate institution type: %w", err)
		}
		if institutionType == nil {
			return nil, nil, fmt.Errorf("invalid institution_type: %s. Valid types: %v",
				institutionTypeName, types.AllInstitutionTypeNames())
		}
		institutionTypeID = institutionType.ID
	}

	// 6. Validate professional type for personal accounts
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

	// 7. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData["password"]), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 8. Create user (NOW all validations have passed)
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

	// 9. Save user to database
	if err := s.repo.CreateUser(user); err != nil {
		return nil, nil, fmt.Errorf("failed to save user: %w", err)
	}

	// ============================================================
	// ✅ SETUP CASBIN POLICIES - ONLY FOR PERSONAL ACCOUNTS
	// ============================================================
	
	if accountTypeName == types.AccountTypePersonalName {
		scope := authdomain.NewPersonalTeamScope(user.ID)

		// 10. Add default policies for the user's personal team
		if err := s.policyManager.AddTeamPolicies(ctx, scope); err != nil {
			log.Printf("⚠️ Failed to add personal team policies: %v", err)
		}

		// 11. Assign account_admin role for user's personal team
		if err := s.roleManager.AssignRole(ctx, scope, user.ID, authdomain.RoleAccountAdmin.String()); err != nil {
			log.Printf("⚠️ Failed to assign personal team admin role: %v", err)
		}

		// 12. Add user to team_members table for their personal team
		if err := s.createPersonalTeamMember(ctx, user.ID); err != nil {
			log.Printf("⚠️ Failed to create personal team member: %v", err)
		}
		
		log.Printf("✅ Personal team created for user: %s", user.ID)
	}

	// ============================================================
	// ✅ CREATE INSTITUTION (if institution account)
	// ============================================================
	if accountTypeName == types.AccountTypeInstitutionName {
		if err := s.createInstitution(ctx, user.ID, userData, institutionTypeID); err != nil {
			return nil, nil, fmt.Errorf("failed to create institution: %w", err)
		}
		log.Printf("✅ Institution created successfully for user: %s", user.ID)
		
		// ✅ Refresh user to get the updated InstitutionID
		updatedUser, err := s.repo.GetUserByID(user.ID)
		if err != nil {
			log.Printf("⚠️ Failed to refresh user: %v", err)
		} else if updatedUser != nil {
			user = updatedUser
			log.Printf("✅ User refreshed with InstitutionID: %v", user.InstitutionID)
		}
	}

	// ============================================================
	// ✅ GENERATE TOKENS
	// ============================================================
	accessToken, refreshToken, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 13. Clean up Redis data
	if err := s.DeleteOTP(ctx, email, "registration"); err != nil {
		log.Printf("Failed to delete OTP: %v", err)
	}
	if err := s.DeleteUserData(ctx, email); err != nil {
		log.Printf("Failed to delete user data: %v", err)
	}

	// 14. Build additional data
	additionalData := map[string]interface{}{
		"user_id":       user.ID,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	// Add institution data if institution account
	if accountTypeName == types.AccountTypeInstitutionName {
		additionalData["institution_name"] = userData["institution_name"]
		additionalData["institution_email"] = userData["institution_email"]
		additionalData["institution_phone"] = userData["institution_phone"]
		additionalData["institution_type"] = userData["institution_type"]

		// Fetch the user again to get the updated institution_id
		updatedUser, err := s.repo.GetUserByID(user.ID)
		if err == nil && updatedUser != nil {
			user = updatedUser
			if user.InstitutionID != nil {
				additionalData["institution_id"] = *user.InstitutionID
			}
		}
	}

	// Add professional type to response
	if user.ProfessionalTypeID != nil {
		additionalData["professional_type_id"] = *user.ProfessionalTypeID
	}
	if userData["professional_type"] != "" {
		additionalData["professional_type"] = userData["professional_type"]
	}

	// 15. Send welcome emails
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

// createPersonalTeamMember creates a team_member record for a user's personal team
func (s *service) createPersonalTeamMember(ctx context.Context, userID string) error {
	// Get personal team type
	teamType, err := s.repo.GetTeamTypeBySlug("personal-team")
	if err != nil {
		return fmt.Errorf("failed to get personal team type: %w", err)
	}
	if teamType == nil {
		return errors.New("personal team type not found")
	}

	// Create team member (user is admin of their own personal team)
	teamMember, err := authdomain.NewPersonalTeamMember(userID, teamType.ID)
	if err != nil {
		return err
	}

	if err := s.repo.CreateTeamMember(teamMember); err != nil {
		return fmt.Errorf("failed to create personal team member: %w", err)
	}

	log.Printf("✅ Personal team member created for user: %s", userID)
	return nil
}

// createInstitution creates an institution and sets up the institution team
func (s *service) createInstitution(ctx context.Context, userID string, userData map[string]string, institutionTypeID string) error {
	institutionName := userData["institution_name"]
	if institutionName == "" {
		return errors.New("institution name is required")
	}

	institutionEmail := userData["institution_email"]
	if institutionEmail == "" {
		return errors.New("institution email is required")
	}

	institutionPhone := userData["institution_phone"]
	if institutionPhone == "" {
		return errors.New("institution phone is required")
	}

	institution, err := authdomain.NewInstitution(
		institutionName,
		institutionEmail,
		institutionPhone,
		institutionTypeID,
	)
	if err != nil {
		return err
	}

	if description := userData["description"]; description != "" {
		institution.Description = description
	}
	if website := userData["website"]; website != "" {
		institution.Website = website
	}

	if err := s.repo.CreateInstitution(institution); err != nil {
		return fmt.Errorf("failed to create institution: %w", err)
	}

	// Update user with institution ID
	if err := s.repo.UpdateUserInstitutionID(userID, &institution.ID); err != nil {
		return fmt.Errorf("failed to update user with institution: %w", err)
	}

	// Get institution team type
	teamType, err := s.repo.GetTeamTypeBySlug("institution-team")
	if err != nil {
		return fmt.Errorf("failed to get institution team type: %w", err)
	}
	if teamType == nil {
		return errors.New("institution team type not found")
	}

	// Create team member (admin)
	teamMember, err := authdomain.NewTeamMember(userID, institution.ID, teamType.ID)
	if err != nil {
		return err
	}

	if err := s.repo.CreateTeamMember(teamMember); err != nil {
		return fmt.Errorf("failed to create team member: %w", err)
	}

	// ============================================================
	// ✅ ASSIGN INSTITUTION ADMIN ROLE VIA CASBIN
	// ============================================================
	scope := authdomain.NewInstitutionTeamScope(institution.ID)
	
	// Add institution policies
	if err := s.policyManager.AddTeamPolicies(ctx, scope); err != nil {
		log.Printf("⚠️ Failed to add institution policies: %v", err)
	}

	// Assign account_admin role
	if err := s.roleManager.AssignRole(ctx, scope, userID, authdomain.RoleAccountAdmin.String()); err != nil {
		log.Printf("⚠️ Failed to assign institution admin role: %v", err)
	}

	log.Printf("✅ Institution team created for institution: %s", institution.ID)
	return nil
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

		// Send KYC welcome
		if err := s.notifSvc.SendInstitutionKYCWelcome(ctx, authdomain.SendInstitutionKYCWelcomeRequest{
			To:               userData["institution_email"],
			AdminName:        user.Name,
			InstitutionName:  userData["institution_name"],
			InstitutionType:  userData["institution_type"],
		}); err != nil {
			errMsg := fmt.Errorf("failed to send institution KYC welcome: %w", err)
			log.Printf("[sendWelcomeEmails] %v", errMsg)
			errors = append(errors, errMsg)
		}

		// Send notification to internal admin
		if err := s.notifSvc.SendNewInstitutionAccountNotification(ctx, authdomain.SendNewInstitutionAccountRegistrationRequest{
			To:                  s.config.NuruOnboardingNoticeEmails.AdminEmail,
			NewAccountAdminName: user.Name,
			InstitutionName:     userData["institution_name"],
			InstitutionType:     userData["institution_type"],
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