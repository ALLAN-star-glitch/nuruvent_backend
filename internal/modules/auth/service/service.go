// internal/modules/auth/service/service.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/domain"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/email"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/queue"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/redis"
	"golang.org/x/crypto/bcrypt"
)

// Service defines the auth service interface
type Service interface {
	// Registration
	RegisterAccount(ctx context.Context, req RegisterRequest) error
	VerifyOTPAndCreateAccount(ctx context.Context, email, otp string) (*models.Account, map[string]interface{}, error)

	// Login
	LoginAccount(ctx context.Context, email, password, ipAddress, userAgent string) (*models.Account, string, error)
	VerifyTwoFactorAndLogin(ctx context.Context, email, otp, ipAddress, userAgent string) (*models.Account, string, string, error)

	// Token management
	GenerateTokens(ctx context.Context, account *models.Account) (string, string, error)
	RefreshTokens(ctx context.Context, refreshToken, userAgent, ip string) (string, string, error)
	RevokeToken(ctx context.Context, refreshToken string) error

	// Password reset
	InitiatePasswordReset(ctx context.Context, email, newPassword string) error
	VerifyResetOTPAndResetPassword(ctx context.Context, email, otp string) error

	// OTP
	GenerateOTP() string
	StoreOTP(email, otp string) error
	GetOTP(email string) (string, error)
	DeleteOTP(email string) error

	// Email sending
	SendOTPEmail(to, name, otp string) error
}

// RegisterRequest holds registration data
type RegisterRequest struct {
	Email       string
	Password    string
	Name        string
	Phone       string
	AccountType string

	// Institution fields (only for institution accounts)
	InstitutionName  string
	InstitutionEmail string
	InstitutionPhone string
	InstitutionType  string
}


type service struct {
	repo         domain.Repository
	config       *config.Config
	queue        *queue.Client
	permService  *authorization.Service
	tokenService TokenService
	emailService *email.EmailService
}

func NewService(
	repo domain.Repository,
	cfg *config.Config,
	queueClient *queue.Client,
	permService *authorization.Service,
	tokenService TokenService,
	emailService *email.EmailService,
) Service {
	return &service{
		repo:         repo,
		config:       cfg,
		queue:        queueClient,
		permService:  permService,
		tokenService: tokenService,
		emailService: emailService,
	}
}

// ================================================
// OTP GENERATION & MANAGEMENT
// ================================================

func (s *service) GenerateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *service) StoreOTP(email, otp string) error {
	key := "otp:" + email
	return redis.Set(key, otp, 5*time.Minute)
}

func (s *service) GetOTP(email string) (string, error) {
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

func (s *service) DeleteOTP(email string) error {
	key := "otp:" + email
	return redis.Delete(key)
}

func (s *service) storeUserData(email string, data map[string]interface{}) error {
	key := "user:data:" + email
	if err := redis.HSet(key, data); err != nil {
		return err
	}
	return redis.Expire(key, 5*time.Minute)
}

func (s *service) getUserData(email string) (map[string]string, error) {
	key := "user:data:" + email
	result, exists, err := redis.HGetAll(key)
	if err != nil {
		return nil, err
	}
	if !exists || len(result) == 0 {
		return nil, errors.New("user data not found")
	}
	return result, nil
}

func (s *service) deleteUserData(email string) error {
	key := "user:data:" + email
	return redis.Delete(key)
}

// ================================================
// REGISTRATION
// ================================================

func (s *service) RegisterAccount(ctx context.Context, req RegisterRequest) error {
	// Check if account exists
	exists, err := s.repo.AccountExistsByEmail(req.Email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email already registered")
	}

	exists, err = s.repo.AccountExistsByPhone(req.Phone)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("phone already registered")
	}

	// Generate OTP
	otp := s.GenerateOTP()

	// Store OTP
	if err := s.StoreOTP(req.Email, otp); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// Store user data
	userData := map[string]interface{}{
		"email":        req.Email,
		"password":     req.Password,
		"name":         req.Name,
		"phone":        req.Phone,
		"account_type": req.AccountType,
	}

	// Add institution fields if provided
	if req.AccountType == "institution" {
		userData["institution_name"] = req.InstitutionName
		userData["institution_email"] = req.InstitutionEmail
		userData["institution_phone"] = req.InstitutionPhone
		userData["institution_type"] = req.InstitutionType
	}

	if err := s.storeUserData(req.Email, userData); err != nil {
		return fmt.Errorf("failed to store user data: %w", err)
	}

	// Send OTP email
	if err := s.SendOTPEmail(req.Email, req.Name, otp); err != nil {
		log.Printf("Failed to send OTP: %v", err)
	}

	return nil
}

func (s *service) SendOTPEmail(to, name, otp string) error {
	return s.emailService.SendVerificationOTP(to, name, otp, "5 minutes", "registration", nil)
}

// ================================================
// OTP VERIFICATION & ACCOUNT CREATION
// ================================================

func (s *service) VerifyOTPAndCreateAccount(ctx context.Context, email, otp string) (*models.Account, map[string]interface{}, error) {
	// Verify OTP
	storedOTP, err := s.GetOTP(email)
	if err != nil {
		return nil, nil, errors.New("invalid or expired OTP")
	}
	if otp != storedOTP {
		return nil, nil, errors.New("invalid OTP")
	}

	// Get user data
	userData, err := s.getUserData(email)
	if err != nil {
		return nil, nil, errors.New("registration data not found")
	}

	// Clean up
	s.DeleteOTP(email)
	s.deleteUserData(email)

	// Create account
	now := time.Now()
	account := &models.Account{
		Email:          email,
		Name:           userData["name"],
		Phone:          userData["phone"],
		EmailVerified:  true,
		EmailVerifiedAt: &now,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Set account type
	accountType, err := s.repo.GetAccountTypeBySlug(userData["account_type"])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get account type: %w", err)
	}
	if accountType == nil {
		return nil, nil, fmt.Errorf("invalid account type: %s", userData["account_type"])
	}
	account.AccountTypeID = accountType.ID

	// If institution, create institution
	var institution *models.Institution
	if userData["account_type"] == "institution" {
		institutionType, err := s.repo.GetInstitutionTypeBySlug(userData["institution_type"])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get institution type: %w", err)
		}
		if institutionType == nil {
			return nil, nil, fmt.Errorf("invalid institution type: %s", userData["institution_type"])
		}

		institution = &models.Institution{
			Name:              userData["institution_name"],
			Email:             userData["institution_email"],
			Phone:             userData["institution_phone"],
			InstitutionTypeID: institutionType.ID,
			IsActive:          true,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		if err := s.repo.CreateInstitution(institution); err != nil {
			return nil, nil, fmt.Errorf("failed to create institution: %w", err)
		}

		// Link institution to account
		account.InstitutionID = &institution.ID

		// Create team member (institution admin)
		teamMember := &models.TeamMember{
			AccountID:     account.ID,
			MemberID:      account.ID,
			Role:          "admin",
			IsActive:      true,
			JoinedAt:      now,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := s.repo.CreateTeamMember(teamMember); err != nil {
			return nil, nil, fmt.Errorf("failed to create team member: %w", err)
		}
	}

	// Save account
	if err := s.repo.CreateAccount(account); err != nil {
		return nil, nil, fmt.Errorf("failed to create account: %w", err)
	}

	// Generate tokens
	accessToken, refreshToken, err := s.GenerateTokens(ctx, account)
	if err != nil {
		return nil, nil, err
	}

	// Send welcome email
	if err := s.emailService.SendWelcome(account.Email, account.Name); err != nil {
		log.Printf("Failed to send welcome email: %v", err)
	}

	// Build result
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

// ================================================
// TOKEN MANAGEMENT
// ================================================

func (s *service) GenerateTokens(ctx context.Context, account *models.Account) (string, string, error) {
	accessToken, err := s.tokenService.GenerateAccessToken(account)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(
		account.ID.String(),
		"",
		"",
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *service) RefreshTokens(ctx context.Context, refreshToken, userAgent, ip string) (string, string, error) {
	newAccessToken, newRefreshToken, err := s.tokenService.RotateRefreshToken(refreshToken, userAgent, ip)
	if err != nil {
		return "", "", err
	}
	return newAccessToken, newRefreshToken, nil
}

func (s *service) RevokeToken(ctx context.Context, refreshToken string) error {
	return s.tokenService.RevokeRefreshToken(refreshToken)
}

// ================================================
// LOGIN
// ================================================

func (s *service) LoginAccount(ctx context.Context, email, password, ipAddress, userAgent string) (*models.Account, string, error) {
	account, err := s.AuthenticateAccountByEmail(email, password)
	if err != nil {
		return nil, "", err
	}

	// Generate 2FA OTP
	otp := s.GenerateOTP()
	if err := s.StoreTwoFactorOTP(account.Email, otp); err != nil {
		return nil, "", fmt.Errorf("failed to store 2FA OTP: %w", err)
	}

	// Send 2FA OTP email
	if err := s.SendTwoFactorOTP(account.Email, account.Name, otp, ipAddress, userAgent); err != nil {
		log.Printf("Failed to send 2FA OTP email: %v", err)
	}

	return account, otp, nil
}

func (s *service) AuthenticateAccountByEmail(email, password string) (*models.Account, error) {
	account, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, errors.New("invalid credentials")
	}

	// TODO: Compare password
	// if err := account.ComparePassword(password); err != nil {
	// 	return nil, errors.New("invalid credentials")
	// }

	return account, nil
}

func (s *service) StoreTwoFactorOTP(email, otp string) error {
	key := fmt.Sprintf("2fa:%s", email)
	return redis.Set(key, otp, 5*time.Minute)
}

func (s *service) GetTwoFactorOTP(email string) (string, error) {
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

func (s *service) DeleteTwoFactorOTP(email string) error {
	key := fmt.Sprintf("2fa:%s", email)
	return redis.Delete(key)
}

func (s *service) SendTwoFactorOTP(to, name, otp, ipAddress, userAgent string) error {
	return s.emailService.SendVerificationOTP(to, name, otp, "5 minutes", "two-factor", map[string]string{
		"ip":         ipAddress,
		"user_agent": userAgent,
	})
}

func (s *service) VerifyTwoFactorAndLogin(ctx context.Context, email, otp, ipAddress, userAgent string) (*models.Account, string, string, error) {
	// Verify OTP
	storedOTP, err := s.GetTwoFactorOTP(email)
	if err != nil {
		return nil, "", "", errors.New("invalid or expired OTP")
	}
	if otp != storedOTP {
		return nil, "", "", errors.New("invalid OTP")
	}

	// Get account
	account, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return nil, "", "", err
	}
	if account == nil {
		return nil, "", "", errors.New("account not found")
	}

	// Delete OTP
	s.DeleteTwoFactorOTP(email)

	// Generate tokens
	accessToken, refreshToken, err := s.GenerateTokens(ctx, account)
	if err != nil {
		return nil, "", "", err
	}

	// Send login notification
	if err := s.SendLoginNotification(account.Email, account.Name, ipAddress, userAgent); err != nil {
		log.Printf("Failed to send login notification: %v", err)
	}

	return account, accessToken, refreshToken, nil
}

func (s *service) SendLoginNotification(to, name, ipAddress, userAgent string) error {
	now := time.Now().Format("2006-01-02 15:04:05 UTC")
	return s.emailService.SendLoginNotification(to, name, now, ipAddress, userAgent)
}

// ================================================
// PASSWORD RESET
// ================================================

func (s *service) InitiatePasswordReset(ctx context.Context, email, newPassword string) error {
	account, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return err
	}
	if account == nil {
		// Don't reveal if account exists
		return nil
	}

	otp := s.GenerateOTP()
	if err := s.StoreResetData(email, otp, newPassword); err != nil {
		return fmt.Errorf("failed to store reset data: %w", err)
	}

	// Send OTP email
	if err := s.SendPasswordResetOTP(email, account.Name, otp); err != nil {
		log.Printf("Failed to send password reset OTP: %v", err)
	}

	return nil
}

func (s *service) StoreResetData(email, otp, newPassword string) error {
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

func (s *service) GetResetData(email string) (map[string]string, error) {
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

func (s *service) DeleteResetData(email string) error {
	key := fmt.Sprintf("reset:%s", email)
	return redis.Delete(key)
}

func (s *service) SendPasswordResetOTP(to, name, otp string) error {
	return s.emailService.SendVerificationOTP(to, name, otp, "5 minutes", "password-reset", nil)
}

// internal/modules/auth/service/service.go

// VerifyResetOTPAndResetPassword verifies OTP and resets password
func (s *service) VerifyResetOTPAndResetPassword(ctx context.Context, email, otp string) error {
	// Get reset data
	data, err := s.GetResetData(email)
	if err != nil {
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

	// Get account
	account, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("account not found")
	}

	// Update password
	// First, hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	account.Password = string(hashedPassword)

	if err := s.repo.UpdateAccount(account); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Clean up
	s.DeleteResetData(email)

	// Send confirmation
	if err := s.SendPasswordResetConfirm(email, account.Name); err != nil {
		log.Printf("Failed to send password reset confirmation: %v", err)
	}

	return nil
}

func (s *service) SendPasswordResetConfirm(to, name string) error {
	return s.emailService.SendPasswordResetConfirm(to, name)
}