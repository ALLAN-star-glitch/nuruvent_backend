package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	authemail "github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/auth_email"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/business/bizservice"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/email"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/queue"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/redis"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type Service struct {
	repo            *Repository
	config          *config.Config
	queue           *queue.Client
	permService     *authorization.Service
	tokenService    *TokenService
	businessService *bizservice.BusinessService
	emailService    *email.EmailService
}

func NewService(
	repo *Repository,
	cfg *config.Config,
	queueClient *queue.Client,
	permService *authorization.Service,
	tokenService *TokenService,
	businessService *bizservice.BusinessService,
	emailService *email.EmailService,
) *Service {
	return &Service{
		repo:            repo,
		config:          cfg,
		queue:           queueClient,
		permService:     permService,
		tokenService:    tokenService,
		businessService: businessService,
		emailService:    emailService,
	}
}

// ================================================
// OTP GENERATION & MANAGEMENT
// ================================================

// GenerateOTP - ALWAYS random
func (s *Service) GenerateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// StoreOTP in Redis with TTL (for registration)
func (s *Service) StoreOTP(email, otp string) error {
	key := "otp:" + email
	if err := redis.Set(key, otp, 5*time.Minute); err != nil {
		return err
	}
	return nil
}

// GetOTP from Redis (for registration)
func (s *Service) GetOTP(email string) (string, error) {
	key := "otp:" + email
	otp, exists, err := redis.Get(key)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New("OTP not found")
	}
	return otp, nil
}

// DeleteOTP from Redis (for registration)
func (s *Service) DeleteOTP(email string) error {
	key := "otp:" + email
	return redis.Delete(key)
}

// StoreTwoFactorOTP stores 2FA OTP in Redis
func (s *Service) StoreTwoFactorOTP(email, otp string) error {
	key := fmt.Sprintf("2fa:%s", email)
	return redis.Set(key, otp, 5*time.Minute)
}

// GetTwoFactorOTP retrieves 2FA OTP from Redis
func (s *Service) GetTwoFactorOTP(email string) (string, error) {
	key := fmt.Sprintf("2fa:%s", email)
	otp, exists, err := redis.Get(key)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New("2FA OTP not found or expired")
	}
	return otp, nil
}

// DeleteTwoFactorOTP removes 2FA OTP from Redis
func (s *Service) DeleteTwoFactorOTP(email string) error {
	key := fmt.Sprintf("2fa:%s", email)
	return redis.Delete(key)
}

// ================================================
// USER REGISTRATION DATA MANAGEMENT
// ================================================

// StoreUserData stores user registration data in Redis
func (s *Service) StoreUserData(email string, data map[string]interface{}) error {
	key := fmt.Sprintf("user:data:%s", email)
	if err := redis.HSet(key, data); err != nil {
		return err
	}
	return redis.Expire(key, 5*time.Minute)
}

// GetUserData gets user registration data from Redis
func (s *Service) GetUserData(email string) (map[string]string, error) {
	key := fmt.Sprintf("user:data:%s", email)
	result, exists, err := redis.HGetAll(key)
	if err != nil {
		return nil, err
	}
	if !exists || len(result) == 0 {
		return nil, errors.New("user data not found")
	}
	return result, nil
}

// DeleteUserData deletes user registration data from Redis
func (s *Service) DeleteUserData(email string) error {
	key := fmt.Sprintf("user:data:%s", email)
	return redis.Delete(key)
}

// FindUserDataByEmail finds user data by email (searches by personal email or business email)
func (s *Service) FindUserDataByEmail(email string) (map[string]string, error) {
	// First try to get by personal email
	data, err := s.GetUserData(email)
	if err == nil {
		return data, nil
	}

	// If not found, check if there's a mapping for this email
	mappingKey := fmt.Sprintf("email:mapping:%s", email)
	mappedEmail, exists, err := redis.Get(mappingKey)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("user data not found")
	}

	// Get user data using the mapped personal email
	return s.GetUserData(mappedEmail)
}

// ================================================
// HELPER METHODS
// ================================================

// isIndividual checks if the user is an individual professional (formal or informal)
func (s *Service) isIndividual(userData map[string]string) bool {
	businessType := userData["business_type"]
	return businessType == "individual_formal" || businessType == "individual_informal" || businessType == "individual"
}

// isFormalIndividual checks if the user is a formal individual professional
func (s *Service) isFormalIndividual(userData map[string]string) bool {
	return userData["business_type"] == "individual_formal"
}

// isOrganization checks if the user is an organization
func (s *Service) isOrganization(userData map[string]string) bool {
	if !s.isBusinessAdmin(userData) {
		return false
	}
	businessType := userData["business_type"]
	// Organization types are anything that's not individual_formal or individual_informal
	return businessType != "individual_formal" && businessType != "individual_informal" && businessType != ""
}

// isBusinessAdmin checks if the user is a business admin
func (s *Service) isBusinessAdmin(userData map[string]string) bool {
	return userData["role"] == string(authorization.RoleBusinessAdmin)
}

// ================================================
// USER CREATION & BUSINESS ONBOARDING
// ================================================


// RegisterUser handles the complete user registration flow
func (s *Service) RegisterUser(ctx context.Context, userData map[string]string) error {
	// Determine user type
	isBusinessAdmin := s.isBusinessAdmin(userData)
	isOrganization := s.isOrganization(userData)
	isFormalIndividual := s.isFormalIndividual(userData)

	// For organizations and formal individuals, use business_email for checks
	primaryEmail := userData["email"]
	if isOrganization || isFormalIndividual {
		if userData["business_email"] != "" {
			primaryEmail = userData["business_email"]
		}
	}

	// Check if user exists using the primary email
	if primaryEmail != "" {
		existingUser, _ := s.repo.GetUserByEmail(primaryEmail)
		if existingUser != nil {
			return errors.New("email already registered")
		}
	}

	// For organizations and formal individuals, also check business email uniqueness
	if isOrganization || isFormalIndividual {
		if userData["business_email"] != "" {
			existingBusiness, _ := s.businessService.GetBusinessByEmail(ctx, userData["business_email"])
			if existingBusiness != nil {
				return errors.New("business email already registered")
			}
		}
	}

	// Check phone if provided
	if userData["phone"] != "" {
		existingPhone, _ := s.repo.GetUserByPhone(userData["phone"])
		if existingPhone != nil {
			return errors.New("phone number already registered")
		}
	}

	// Generate OTP
	otp := s.GenerateOTP()

	// Store user data - use business email as key if personal email is empty
	storeKey := userData["email"]
	if storeKey == "" && userData["business_email"] != "" {
		storeKey = userData["business_email"]
	}
	if storeKey == "" {
		return errors.New("no email provided for storing user data")
	}

	userDataInterface := make(map[string]any)
	for key, value := range userData {
		userDataInterface[key] = value
	}
	if err := s.StoreUserData(storeKey, userDataInterface); err != nil {
		return fmt.Errorf("failed to store user data: %w", err)
	}

	// Determine where to send OTP
	if isBusinessAdmin && isOrganization {
		// Organization - OTP goes to business email
		businessEmail := userData["business_email"]
		if businessEmail == "" {
			return errors.New("business email is required for organizations")
		}

		// Store mapping: business_email -> storeKey
		mappingKey := fmt.Sprintf("email:mapping:%s", businessEmail)
		if err := redis.Set(mappingKey, storeKey, 5*time.Minute); err != nil {
			log.Printf("Failed to store email mapping: %v", err)
		}

		// Store OTP with business email as key
		if err := s.StoreOTP(businessEmail, otp); err != nil {
			return fmt.Errorf("failed to store OTP: %w", err)
		}
		// Send OTP to business email
		businessName := userData["business_name"]
		if businessName == "" {
			businessName = userData["name"]
		}
		if err := s.SendOTPEmail(businessEmail, businessName, otp); err != nil {
			log.Printf("Failed to send OTP email: %v", err)
		}
	} else if isBusinessAdmin && isFormalIndividual && userData["business_email"] != "" {
		// Formal Individual with business email - OTP goes to business email
		businessEmail := userData["business_email"]

		// Store mapping: business_email -> storeKey
		mappingKey := fmt.Sprintf("email:mapping:%s", businessEmail)
		if err := redis.Set(mappingKey, storeKey, 5*time.Minute); err != nil {
			log.Printf("Failed to store email mapping: %v", err)
		}

		// Store OTP with business email as key
		if err := s.StoreOTP(businessEmail, otp); err != nil {
			return fmt.Errorf("failed to store OTP: %w", err)
		}
		// Send OTP to business email
		businessName := userData["business_name"]
		if businessName == "" {
			businessName = userData["name"]
		}
		if err := s.SendOTPEmail(businessEmail, businessName, otp); err != nil {
			log.Printf("Failed to send OTP email: %v", err)
		}
	} else {
		// Attendee OR Informal Individual OR Formal Individual without business email
		// OTP goes to personal email
		if userData["email"] == "" {
			return errors.New("email is required for attendees and informal individuals")
		}
		if err := s.StoreOTP(userData["email"], otp); err != nil {
			return fmt.Errorf("failed to store OTP: %w", err)
		}
		// Send OTP to personal email
		if err := s.SendOTPEmail(userData["email"], userData["name"], otp); err != nil {
			log.Printf("Failed to send OTP email: %v", err)
		}
	}

	return nil
}

// VerifyOTPAndCreateUser handles OTP verification and user creation
func (s *Service) VerifyOTPAndCreateUser(ctx context.Context, email, otp string) (*models.User, map[string]interface{}, error) {
	// Verify OTP
	storedOTP, err := s.GetOTP(email)
	if err != nil {
		return nil, nil, errors.New("invalid or expired OTP")
	}

	if otp != storedOTP {
		return nil, nil, errors.New("invalid OTP")
	}

	// Find user data (could be personal email or business email)
	userDataMap, err := s.FindUserDataByEmail(email)
	if err != nil {
		return nil, nil, errors.New("registration data not found")
	}

	// Clean up OTP
	s.DeleteOTP(email)

	// Clean up email mapping if it exists
	mappingKey := fmt.Sprintf("email:mapping:%s", email)
	redis.Delete(mappingKey)

	// Clean up user data - use business email if personal email is empty
	personalEmail := userDataMap["email"]
	if personalEmail == "" && userDataMap["business_email"] != "" {
		personalEmail = userDataMap["business_email"]
	}
	if personalEmail != "" {
		s.DeleteUserData(personalEmail)
	}

	// Check if this is a business_admin with business
	isBusinessAdmin := s.isBusinessAdmin(userDataMap)
	isIndividual := s.isIndividual(userDataMap)

	if isBusinessAdmin {
		// Create user with business (Individual Professional or Organization)
		user, business, result, err := s.CreateUserWithBusiness(ctx, userDataMap)
		if err != nil {
			return nil, nil, err
		}
		result["user"] = user
		result["business"] = business

		// If Individual Professional, log it
		if isIndividual {
			log.Printf("Individual Professional business created: %s", business.ID)
		}

		return user, result, nil
	}

	// Create user without business (attendee)
	user, err := s.CreateUserOnly(ctx, userDataMap)
	if err != nil {
		return nil, nil, err
	}

	// Generate tokens
	accessToken, refreshToken, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	result := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    int64(s.config.JWT.AccessExpiration.Seconds()),
		"user":          user,
	}

	// Send welcome email
	if err := s.SendWelcomeEmail(user.Email, user.Name); err != nil {
		log.Printf("Failed to send welcome email: %v", err)
	}

	return user, result, nil
}

// CreateUserOnly creates a user without creating a business
func (s *Service) CreateUserOnly(ctx context.Context, userData map[string]string) (*models.User, error) {
	now := time.Now()
	
	// For organizations and formal individuals, use business_email as the primary email
	email := userData["email"]
	if email == "" && userData["business_email"] != "" {
		email = userData["business_email"]
	}
	if email == "" {
		return nil, errors.New("email is required")
	}

	// For organizations, name can be empty, use business_name as fallback
	name := userData["name"]
	if name == "" && userData["business_name"] != "" {
		name = userData["business_name"]
	}

	// Phone can be empty for organizations - will be NULL in database
	phone := userData["phone"]

	user := &models.User{
		Email:           email,
		Phone:           phone, // Empty string will be stored as NULL
		PasswordHash:    "",
		Name:            name,
		Role:            userData["role"],
		IsVerified:      true,
		IsEmailVerified: true,
		VerifiedAt:      &now,
		EmailVerifiedAt: &now,
		LastActiveAt:    &now,
	}

	// Hash password
	if err := user.HashPassword(userData["password"]); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign attendee role (all users get this)
	if err := s.permService.AssignAttendeeRole(ctx, user.ID.String()); err != nil {
		log.Printf("Failed to assign attendee role: %v", err)
	}

	return user, nil
}

// CreateUserWithBusiness creates a user and immediately creates their business
func (s *Service) CreateUserWithBusiness(ctx context.Context, userData map[string]string) (*models.User, *models.Business, map[string]interface{}, error) {
	// 1. Create user
	user, err := s.CreateUserOnly(ctx, userData)
	if err != nil {
		return nil, nil, nil, err
	}

	// 2. CRITICAL: Create a fresh business data map with required fields
	businessData := make(map[string]string)
	
	// Copy all user data
	for k, v := range userData {
		businessData[k] = v
	}
	
	// Explicitly set email and name from user object for ALL business types
	businessData["email"] = user.Email
	businessData["name"] = user.Name

	log.Printf("Business data prepared - email: '%s', name: '%s', business_type: '%s'",
		businessData["email"], businessData["name"], businessData["business_type"])

	// 3. Validate business type exists
	businessTypeName := userData["business_type"]
	if businessTypeName != "" {
		businessType, err := s.businessService.GetBusinessTypeByName(ctx, businessTypeName)
		if err != nil {
			log.Printf("Warning: Failed to get business type by name: %v", err)
		} else if businessType == nil {
			if businessTypeIDStr, ok := userData["business_type_id"]; ok && businessTypeIDStr != "" {
				businessTypeID, err := uuid.Parse(businessTypeIDStr)
				if err == nil {
					businessType, err = s.businessService.GetBusinessTypeByID(ctx, businessTypeID)
					if err != nil {
						log.Printf("Warning: Failed to get business type by ID: %v", err)
					}
				}
			}
		}
		if businessType != nil {
			log.Printf("Business type verified: %s (ID: %s)", businessType.Name, businessType.ID)
		}
	}

	// 4. Create business using business service with the new map
	business, err := s.businessService.CreateBusinessFromRegistration(ctx, user.ID, businessData)
	if err != nil {
		return user, nil, nil, fmt.Errorf("user created but business creation failed: %w", err)
	}

	// 5. Update user with business ID
	user.BusinessID = &business.ID
	if err := s.repo.UpdateUser(user); err != nil {
		log.Printf("Failed to update user with business ID: %v", err)
	}

	// 6. Generate tokens
	accessToken, refreshToken, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return user, business, nil, err
	}

	// 7. Send welcome email based on user type
	userBusinessType := userData["business_type"]
	
	// Determine welcome email destination
	var welcomeEmailTo string

	if userBusinessType == "individual_formal" || userData["business_type"] == "organization" || s.isOrganization(userData) {
		// Formal Individual or Organization - send to business email
		if business.Email != "" {
			welcomeEmailTo = business.Email
		} else if userData["business_email"] != "" {
			welcomeEmailTo = userData["business_email"]
		} else {
			welcomeEmailTo = user.Email
		}
		log.Printf("Formal entity - sending welcome to business email: %s", welcomeEmailTo)
	} else {
		// Informal Individual or Attendee - send to personal email
		welcomeEmailTo = user.Email
		log.Printf("Individual - sending welcome to personal email: %s", welcomeEmailTo)
	}

	if welcomeEmailTo != "" {
		businessName := business.Name
		if businessName == "" {
			businessName = user.Name
		}

		// Determine which welcome template to use
		isIndividualProfessional := userBusinessType == "individual" ||
			userBusinessType == "individual_formal" ||
			userBusinessType == "individual_informal"

		if isIndividualProfessional {
			// Send individual professional welcome email
			if err := s.SendIndividualProfessionalWelcomeEmail(welcomeEmailTo, user.Name); err != nil {
				log.Printf("Failed to send individual professional welcome email: %v", err)
			}
		} else if userBusinessType != "" && !isIndividualProfessional {
			// Organization - send business welcome email
			if err := s.SendBusinessWelcomeEmail(welcomeEmailTo, businessName, user.Name); err != nil {
				log.Printf("Failed to send business welcome email: %v", err)
			}
		} else {
			// Fallback to regular welcome email
			if err := s.SendWelcomeEmail(welcomeEmailTo, user.Name); err != nil {
				log.Printf("Failed to send welcome email: %v", err)
			}
		}
	}

	// 8. Build result
	result := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    int64(s.config.JWT.AccessExpiration.Seconds()),
		"business_name": business.Name,
		"business_id":   business.ID,
		"user":          user,
		"business":      business,
	}

	log.Printf("Business created with ID: %s, email: %s, email verified: %v",
		business.ID, business.Email, business.IsEmailVerified)

	return user, business, result, nil
}

// ================================================
// TOKEN MANAGEMENT
// ================================================

// GenerateTokens generates access and refresh tokens for a user
func (s *Service) GenerateTokens(ctx context.Context, user *models.User) (string, string, error) {
	accessToken, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(
		user.ID.String(),
		"", // User-Agent will be set by handler
		"", // IP will be set by handler
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// RefreshTokens refreshes access token using refresh token
func (s *Service) RefreshTokens(ctx context.Context, refreshToken, userAgent, ip string) (string, string, error) {
	return s.tokenService.RotateRefreshToken(refreshToken, userAgent, ip)
}

// RevokeToken revokes a refresh token
func (s *Service) RevokeToken(ctx context.Context, refreshToken string) error {
	return s.tokenService.RevokeRefreshToken(refreshToken)
}

// ================================================
// AUTHENTICATION METHODS
// ================================================

// LoginUser authenticates a user and initiates 2FA
func (s *Service) LoginUser(ctx context.Context, email, password, ipAddress, userAgent string) (*models.User, string, error) {
	user, err := s.AuthenticateUserByEmail(email, password)
	if err != nil {
		return nil, "", err
	}

	// Generate 2FA OTP
	otp := s.GenerateOTP()
	if err := s.StoreTwoFactorOTP(user.Email, otp); err != nil {
		return nil, "", fmt.Errorf("failed to store 2FA OTP: %w", err)
	}

	// Send 2FA OTP email with IP and User-Agent
	if err := s.SendTwoFactorOTP(user.Email, user.Name, otp, ipAddress, userAgent); err != nil {
		log.Printf("Failed to send 2FA OTP email: %v", err)
	}

	return user, otp, nil
}

// VerifyTwoFactorAndLogin verifies 2FA OTP and completes login
func (s *Service) VerifyTwoFactorAndLogin(ctx context.Context, email, otp, ipAddress, userAgent string) (*models.User, string, string, error) {
	// Verify OTP
	storedOTP, err := s.GetTwoFactorOTP(email)
	if err != nil {
		return nil, "", "", errors.New("invalid or expired OTP")
	}

	if otp != storedOTP {
		return nil, "", "", errors.New("invalid OTP")
	}

	// Get user
	user, err := s.repo.GetUserByEmail(email)
	if err != nil || user == nil {
		return nil, "", "", errors.New("user not found")
	}

	// Delete OTP
	s.DeleteTwoFactorOTP(email)

	// Generate tokens
	accessToken, refreshToken, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	// Send login notification with IP and User-Agent
	if err := s.SendLoginNotification(user.Email, user.Name, ipAddress, userAgent); err != nil {
		log.Printf("Failed to send login notification: %v", err)
	}

	return user, accessToken, refreshToken, nil
}

// AuthenticateUserByEmail authenticates a user by email
func (s *Service) AuthenticateUserByEmail(email, password string) (*models.User, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := user.ComparePassword(password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := s.repo.UpdateUser(user); err != nil {
		log.Printf("Failed to update last login: %v", err)
	}

	return user, nil
}

// AuthenticateUserByPhone authenticates a user by phone number
func (s *Service) AuthenticateUserByPhone(phone, password string) (*models.User, error) {
	user, err := s.repo.GetUserByPhone(phone)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := user.ComparePassword(password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := s.repo.UpdateUser(user); err != nil {
		log.Printf("Failed to update last login: %v", err)
	}

	return user, nil
}

// ================================================
// PASSWORD RESET METHODS
// ================================================

// InitiatePasswordReset initiates password reset flow
func (s *Service) InitiatePasswordReset(ctx context.Context, email, newPassword string) error {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		// Don't reveal if user exists
		return nil
	}

	otp := s.GenerateOTP()
	if err := s.StoreResetData(email, otp, newPassword); err != nil {
		return fmt.Errorf("failed to store reset data: %w", err)
	}

	if err := s.SendPasswordResetOTP(email, user.Name, otp); err != nil {
		log.Printf("Failed to send password reset OTP: %v", err)
	}

	return nil
}

// VerifyResetOTPAndResetPassword verifies OTP and resets password
func (s *Service) VerifyResetOTPAndResetPassword(ctx context.Context, email, otp string) error {
	data, err := s.GetResetData(email)
	if err != nil {
		return errors.New("invalid or expired OTP")
	}

	if len(data) == 0 {
		return errors.New("invalid or expired OTP")
	}

	storedOTP, ok := data["otp"]
	if !ok || otp != storedOTP {
		return errors.New("invalid OTP")
	}

	newPassword, ok := data["new_password"]
	if !ok {
		return errors.New("invalid reset data")
	}

	user, err := s.repo.GetUserByEmail(email)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	if err := user.HashPassword(newPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.DeleteResetData(email)

	if err := s.SendPasswordResetConfirm(email, user.Name); err != nil {
		log.Printf("Failed to send password reset confirmation: %v", err)
	}

	return nil
}

// StoreResetData stores password reset data in Redis
func (s *Service) StoreResetData(email, otp, newPassword string) error {
	key := fmt.Sprintf("reset:%s", email)
	data := map[string]interface{}{
		"otp":          otp,
		"new_password": newPassword,
	}
	if err := redis.HSet(key, data); err != nil {
		return err
	}
	return redis.Expire(key, 5*time.Minute)
}

// GetResetData retrieves password reset data from Redis
func (s *Service) GetResetData(email string) (map[string]string, error) {
	key := fmt.Sprintf("reset:%s", email)
	result, exists, err := redis.HGetAll(key)
	if err != nil {
		return nil, err
	}
	if !exists || len(result) == 0 {
		return nil, errors.New("reset data not found")
	}
	return result, nil
}

// DeleteResetData deletes password reset data from Redis
func (s *Service) DeleteResetData(email string) error {
	key := fmt.Sprintf("reset:%s", email)
	return redis.Delete(key)
}

// ================================================
// EMAIL SEND METHODS (with fallback)
// ================================================

// SendVerificationOTP sends verification OTP via queue with fallback
func (s *Service) SendVerificationOTP(to, name, otp, expires string, purpose authemail.VerificationPurpose, meta map[string]string) error {
	// Try queue first
	task := authemail.VerificationOTPTask{
		To:      to,
		Name:    name,
		OTP:     otp,
		Expires: expires,
		Purpose: purpose,
		Meta:    meta,
	}
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	queueErr := s.queue.Enqueue(authemail.TypeVerificationOTP, payload, asynq.MaxRetry(3), asynq.Timeout(30*time.Second))
	if queueErr == nil {
		return nil
	}

	// Fallback to sync
	log.Printf("Queue failed for verification OTP, falling back to sync: %v", queueErr)
	return s.emailService.SendVerificationOTP(to, name, otp, expires, string(purpose), meta)
}

// SendOTPEmail sends registration OTP
func (s *Service) SendOTPEmail(to, name, otp string) error {
	return s.SendVerificationOTP(to, name, otp, "5 minutes", authemail.PurposeRegistration, nil)
}

// SendWelcomeEmail sends welcome email
func (s *Service) SendWelcomeEmail(to, name string) error {
	// Try queue first
	task := authemail.WelcomeTask{
		To:   to,
		Name: name,
	}
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	queueErr := s.queue.Enqueue(authemail.TypeWelcome, payload, asynq.MaxRetry(3), asynq.Timeout(30*time.Second))
	if queueErr == nil {
		return nil
	}

	// Fallback to sync
	log.Printf("Queue failed for welcome email, falling back to sync: %v", queueErr)
	return s.emailService.SendWelcome(to, name)
}

// SendIndividualProfessionalWelcomeEmail sends individual professional welcome email
func (s *Service) SendIndividualProfessionalWelcomeEmail(to, name string) error {
	task := authemail.IndividualProfessionalWelcomeEmailTask{
		To:   to,
		Name: name,
	}

	payload, err := task.Payload()
	if err != nil {
		return err
	}

	queueErr := s.queue.Enqueue(authemail.TypeEmailIndividualProfessionalWelcome, payload, asynq.MaxRetry(3), asynq.Timeout(30*time.Second))

	if queueErr == nil {
		return nil
	}

	log.Printf("Worker failed to send email %s to user %s: Falling back to sync %v", to, name, queueErr)

	return s.emailService.SendIndividualProfessionalWelcome(to, name)
}

// SendBusinessWelcomeEmail sends business welcome email
func (s *Service) SendBusinessWelcomeEmail(to, businessName, ownerName string) error {
	// Try queue first
	task := authemail.BusinessWelcomeTask{
		To:           to,
		BusinessName: businessName,
		OwnerName:    ownerName,
	}
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	queueErr := s.queue.Enqueue(authemail.TypeBusinessWelcome, payload, asynq.MaxRetry(3), asynq.Timeout(30*time.Second))
	if queueErr == nil {
		return nil
	}

	// Fallback to sync
	log.Printf("Queue failed for business welcome email, falling back to sync: %v", queueErr)
	return s.emailService.SendBusinessWelcome(to, businessName, ownerName)
}

// SendPasswordResetOTP sends password reset OTP
func (s *Service) SendPasswordResetOTP(to, name, otp string) error {
	return s.SendVerificationOTP(to, name, otp, "5 minutes", authemail.PurposePasswordReset, nil)
}

// SendPasswordResetConfirm sends password reset confirmation
func (s *Service) SendPasswordResetConfirm(to, name string) error {
	// Try queue first
	task := authemail.PasswordResetConfirmTask{
		To:   to,
		Name: name,
	}
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	queueErr := s.queue.Enqueue(authemail.TypePasswordResetConfirm, payload, asynq.MaxRetry(3), asynq.Timeout(30*time.Second))
	if queueErr == nil {
		return nil
	}

	// Fallback to sync
	log.Printf("Queue failed for password reset confirm, falling back to sync: %v", queueErr)
	return s.emailService.SendPasswordResetConfirm(to, name)
}

// SendTwoFactorOTP sends 2FA OTP
func (s *Service) SendTwoFactorOTP(to, name, otp, ipAddress, userAgent string) error {
	// Store IP and User-Agent for the 2FA email
	// They will be used in the next step (login notification)
	return s.SendVerificationOTP(to, name, otp, "5 minutes", authemail.PurposeTwoFactor, nil)
}

// SendLoginNotification sends login notification
func (s *Service) SendLoginNotification(to, name, ipAddress, userAgent string) error {
	now := time.Now().Format("2006-01-02 15:04:05 UTC")

	// Try queue first
	task := authemail.LoginNotificationTask{
		To:        to,
		Name:      name,
		Time:      now,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	queueErr := s.queue.Enqueue(authemail.TypeLoginNotification, payload, asynq.MaxRetry(3), asynq.Timeout(30*time.Second))
	if queueErr == nil {
		return nil
	}

	// Fallback to sync
	log.Printf("Queue failed for login notification, falling back to sync: %v", queueErr)
	return s.emailService.SendLoginNotification(to, name, now, ipAddress, userAgent)
}

// ================================================
// USER INFO & ADMIN METHODS
// ================================================

// BuildUserInfo builds user info for response
func (s *Service) BuildUserInfo(user *models.User) UserInfo {
	info := UserInfo{
		ID:    user.ID.String(),
		Email: user.Email,
		Phone: user.Phone,
		Name:  user.Name,
		Role:  user.Role,
	}

	if user.BusinessID != nil {
		businessID := user.BusinessID.String()
		info.BusinessID = &businessID
	}

	return info
}

// GetUserByIDWithBusiness gets a user with their business
func (s *Service) GetUserByIDWithBusiness(id string) (*models.User, error) {
	return s.repo.GetUserByID(id)
}

// GetAllUsers gets all users with pagination
func (s *Service) GetAllUsers(page, pageSize int, search string) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.GetAllUsers(pageSize, offset, search)
}

// GetUserStats gets user statistics
func (s *Service) GetUserStats() (map[string]interface{}, error) {
	return s.repo.GetUserStats()
}

// ValidateToken validates JWT token
func (s *Service) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWT.Secret), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", errors.New("invalid token claims")
		}
		return userID, nil
	}
	return "", errors.New("invalid token")
}

// ================================================
// RESPONSE MODELS
// ================================================

// UserInfo struct
type UserInfo struct {
	ID         string  `json:"id"`
	Email      string  `json:"email"`
	Phone      string  `json:"phone,omitempty"`
	Name       string  `json:"name"`
	Role       string  `json:"role"`
	Avatar     string  `json:"avatar,omitempty"`
	BusinessID *string `json:"business_id,omitempty"`
}