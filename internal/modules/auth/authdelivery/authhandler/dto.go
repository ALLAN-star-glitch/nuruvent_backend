package authhandler

import (
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// ============================================================
// REQUEST DTOS
// ============================================================

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	AccountType string `json:"account_type"`

	// Institution fields (only for institution accounts)
	InstitutionName  string `json:"institution_name,omitempty"`
	InstitutionEmail string `json:"institution_email,omitempty"`
	InstitutionPhone string `json:"institution_phone,omitempty"`
	InstitutionType  string `json:"institution_type,omitempty"`
}

// VerifyOTPRequest represents the OTP verification request
type VerifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

// ResendOTPRequest represents the resend OTP request
type ResendOTPRequest struct {
	Email string `json:"email"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// VerifyTwoFactorOTPRequest represents the 2FA verification request
type VerifyTwoFactorOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest represents the logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ForgotPasswordRequest represents the forgot password request
type ForgotPasswordRequest struct {
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
}

// VerifyResetOTPRequest represents the reset OTP verification request
type VerifyResetOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

// ============================================================
// RESPONSE DTOS
// ============================================================

// TokenResponse represents the token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
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

// AccountResponse represents the account response
type AccountResponse struct {
	ID             string     `json:"id"`
	Slug           string     `json:"slug"`
	Name           string     `json:"name"`
	DisplayName    string     `json:"display_name,omitempty"`
	Email          string     `json:"email"`
	Phone          string     `json:"phone"`
	AccountType    string     `json:"account_type"`
	AccountTypeID  string     `json:"account_type_id"`
	EmailVerified  bool       `json:"email_verified"`
	IdentityVerified bool     `json:"identity_verified"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	InstitutionID  *string    `json:"institution_id,omitempty"`
}

// InstitutionResponse represents the institution response
type InstitutionResponse struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Description string `json:"description,omitempty"`
	Logo        string `json:"logo,omitempty"`
	Website     string `json:"website,omitempty"`
	IsActive    bool   `json:"is_active"`
}

// AuthResponse represents the auth response
type AuthResponse struct {
	TokenResponse TokenResponse      `json:"token"`
	Account       AccountResponse    `json:"account"`
	Institution   *InstitutionResponse `json:"institution,omitempty"`
}

// ============================================================
// RESPONSE BUILDERS (Mappers)
// ============================================================

// NewAccountResponse converts authdomain.Account to AccountResponse
func NewAccountResponse(account *authdomain.Account) AccountResponse {
	if account == nil {
		return AccountResponse{}
	}

	// Get account type name
	accountTypeName := ""
	if account.AccountTypeID != "" {
		// We could fetch the name here if needed
	}

	return AccountResponse{
		ID:              account.ID,
		Slug:            account.Slug,
		Name:            account.Name,
		DisplayName:     account.DisplayName,
		Email:           account.Email,
		Phone:           account.Phone,
		AccountType:     accountTypeName,
		AccountTypeID:   account.AccountTypeID,
		EmailVerified:   account.EmailVerified,
		IdentityVerified: account.IdentityVerified,
		IsActive:        account.IsActive,
		CreatedAt:       account.CreatedAt,
		UpdatedAt:       account.UpdatedAt,
		InstitutionID:   account.InstitutionID,
	}
}

// NewInstitutionResponse converts authdomain.Institution to InstitutionResponse
func NewInstitutionResponse(institution *authdomain.Institution) *InstitutionResponse {
	if institution == nil {
		return nil
	}

	return &InstitutionResponse{
		ID:          institution.ID,
		Slug:        institution.Slug,
		Name:        institution.Name,
		DisplayName: institution.DisplayName,
		Email:       institution.Email,
		Phone:       institution.Phone,
		Description: institution.Description,
		Logo:        institution.Logo,
		Website:     institution.Website,
		IsActive:    institution.IsActive,
	}
}