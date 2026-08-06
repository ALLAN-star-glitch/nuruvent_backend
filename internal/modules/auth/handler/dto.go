// internal/modules/auth/handler/dto.go

package handler

import (
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/google/uuid"
)

// ================================================
// REQUEST DTOs
// ================================================

// RegisterRequest represents the registration request
type RegisterRequest struct {
	// Personal fields (for all users)
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"SecurePass123!"`
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Phone    string `json:"phone" binding:"required" example:"+254712345678"`
	
	// Account type: personal or institution
	AccountType string `json:"account_type" binding:"required,oneof=personal institution" example:"personal"`
	
	// Institution fields (only for institution accounts)
	InstitutionName  string `json:"institution_name,omitempty" example:"Nuruvent Training Institute"`
	InstitutionEmail string `json:"institution_email,omitempty" example:"info@nuruventinstitute.com"`
	InstitutionPhone string `json:"institution_phone,omitempty" example:"+254745678901"`
	InstitutionType  string `json:"institution_type,omitempty" example:"training_institute"`
}

// VerifyOTPRequest represents the OTP verification request
type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
	OTP   string `json:"otp" binding:"required,len=6" example:"123456"`
}

// ResendOTPRequest represents the OTP resend request
type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"SecurePass123!"`
}

// VerifyTwoFactorOTPRequest represents the 2FA OTP verification request
type VerifyTwoFactorOTPRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
	OTP   string `json:"otp" binding:"required,len=6" example:"123456"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"abc123xyz789..."`
}

// LogoutRequest represents the logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" example:"abc123xyz789..."`
}

// ForgotPasswordRequest represents the forgot password request
type ForgotPasswordRequest struct {
	Email       string `json:"email" binding:"required,email" example:"john@example.com"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"SecurePass123!"`
}

// VerifyResetOTPRequest represents the reset OTP verification request
type VerifyResetOTPRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
	OTP   string `json:"otp" binding:"required,len=6" example:"123456"`
}

// ChangePasswordRequest represents the change password request (authenticated)
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"OldPass123!"`
	NewPassword     string `json:"new_password" binding:"required,min=8" example:"NewPass123!"`
}

// ================================================
// RESPONSE DTOs
// ================================================

// AccountResponse represents the account in responses
type AccountResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Phone       string    `json:"phone,omitempty"`
	AccountType string    `json:"account_type"`
	Slug        string    `json:"slug"`
	DisplayName string    `json:"display_name,omitempty"`
	IsVerified  bool      `json:"is_verified"`
	CreatedAt   time.Time `json:"created_at"`
}

// InstitutionResponse represents the institution in responses
type InstitutionResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone,omitempty"`
	Type        string    `json:"type"`
	Slug        string    `json:"slug"`
	DisplayName string    `json:"display_name,omitempty"`
	Logo        string    `json:"logo,omitempty"`
	Website     string    `json:"website,omitempty"`
	Description string    `json:"description,omitempty"`
	Address     string    `json:"address,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// TokenResponse represents the token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// AuthResponse represents the full authentication response
type AuthResponse struct {
	TokenResponse
	Account     AccountResponse     `json:"account"`
	Institution *InstitutionResponse `json:"institution,omitempty"`
}

// OTPResponse represents the OTP response
type OTPResponse struct {
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
	Message   string    `json:"message"`
}

// TwoFactorResponse represents the 2FA response
type TwoFactorResponse struct {
	Requires2FA bool   `json:"requires_2fa"`
	Email       string `json:"email"`
	ExpiresIn   int    `json:"expires_in"`
}

// PasswordResetResponse represents the password reset response
type PasswordResetResponse struct {
	Message   string `json:"message"`
	ExpiresIn int    `json:"expires_in"`
}

// ================================================
// BUILDER FUNCTIONS
// ================================================

// NewAccountResponse creates an account response from models.Account
func NewAccountResponse(account *models.Account) AccountResponse {
	resp := AccountResponse{
		ID:          account.ID.String(),
		Email:       account.Email,
		Name:        account.Name,
		Phone:       account.Phone,
		Slug:        account.Slug,
		DisplayName: account.DisplayName,
		IsVerified:  account.EmailVerified,
		CreatedAt:   account.CreatedAt,
	}

	// Check if AccountType is loaded (ID is not nil)
	if account.AccountType.ID != uuid.Nil {
		resp.AccountType = account.AccountType.Slug
	}

	return resp
}

// NewInstitutionResponse creates an institution response from models.Institution
func NewInstitutionResponse(institution *models.Institution) *InstitutionResponse {
	if institution == nil {
		return nil
	}

	resp := &InstitutionResponse{
		ID:          institution.ID.String(),
		Name:        institution.Name,
		Email:       institution.Email,
		Phone:       institution.Phone,
		Slug:        institution.Slug,
		DisplayName: institution.DisplayName,
		Logo:        institution.Logo,
		Website:     institution.Website,
		Description: institution.Description,
		Address:     institution.Address,
		IsActive:    institution.IsActive,
		CreatedAt:   institution.CreatedAt,
	}

	// Check if InstitutionType is loaded (ID is not nil)
	if institution.InstitutionType.ID != uuid.Nil {
		resp.Type = institution.InstitutionType.Slug
	}

	return resp
}

// NewAuthResponse creates a full auth response
func NewAuthResponse(
	accessToken, refreshToken string,
	expiresIn int64,
	account *models.Account,
	institution *models.Institution,
) AuthResponse {
	resp := AuthResponse{
		TokenResponse: TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    expiresIn,
		},
		Account: NewAccountResponse(account),
	}

	if institution != nil {
		resp.Institution = NewInstitutionResponse(institution)
	}

	return resp
}