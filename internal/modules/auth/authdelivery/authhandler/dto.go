package authhandler

import (
	"context"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"
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

	// Professional type (for personal accounts)
	ProfessionalType string `json:"professional_type,omitempty"`

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
	Email   string `json:"email" validate:"required,email"`
	Purpose string `json:"purpose" validate:"required,oneof=registration two_factor password_reset email_change phone_change"`
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

// UserResponse represents the user response
type UserResponse struct {
	ID                 string     `json:"id"`
	Slug               string     `json:"slug"`
	Name               string     `json:"name"`
	DisplayName        string     `json:"display_name,omitempty"`
	Email              string     `json:"email"`
	Phone              string     `json:"phone"`
	AccountType        string     `json:"account_type"`
	AccountTypeID      string     `json:"account_type_id"`
	ProfessionalTypeID *string    `json:"professional_type_id,omitempty"`
	ProfessionalType   string     `json:"professional_type,omitempty"`
	EmailVerified      bool       `json:"email_verified"`
	IdentityVerified   bool       `json:"identity_verified"`
	IsActive           bool       `json:"is_active"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	InstitutionID      *string    `json:"institution_id,omitempty"`
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
	TokenResponse TokenResponse        `json:"token"`
	User          UserResponse         `json:"user"`
	Institution   *InstitutionResponse `json:"institution,omitempty"`
}

// ============================================================
// RESPONSE BUILDERS (Mappers)
// ============================================================

// NewUserResponse converts authdomain.User to UserResponse
// Fetches account type and professional type names from database
func NewUserResponse(user *authdomain.User, svc service.Service, ctx context.Context) UserResponse {
	if user == nil {
		return UserResponse{}
	}

	// Fetch account type name from database
	accountType := ""
	if user.AccountTypeID != "" {
		accountTypeObj, err := svc.GetAccountTypeByID(ctx, user.AccountTypeID)
		if err == nil && accountTypeObj != nil {
			accountType = accountTypeObj.Name // e.g., "account_type_personal"
		}
	}

	// Fetch professional type name from database
	professionalType := ""
	if user.ProfessionalTypeID != nil {
		professionalTypeObj, err := svc.GetProfessionalTypeByID(ctx, *user.ProfessionalTypeID)
		if err == nil && professionalTypeObj != nil {
			professionalType = professionalTypeObj.Name // e.g., "professional_type_trainer"
		}
	}

	return UserResponse{
		ID:                 user.ID,
		Slug:               user.Slug,
		Name:               user.Name,
		DisplayName:        user.DisplayName,
		Email:              user.Email,
		Phone:              user.Phone,
		AccountType:        accountType,
		AccountTypeID:      user.AccountTypeID,
		ProfessionalTypeID: user.ProfessionalTypeID,
		ProfessionalType:   professionalType,
		EmailVerified:      user.EmailVerified,
		IdentityVerified:   user.IdentityVerified,
		IsActive:           user.IsActive,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
		InstitutionID:      user.InstitutionID,
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

// NewAuthResponse creates a complete auth response
func NewAuthResponse(user *authdomain.User, accessToken, refreshToken string, institution *authdomain.Institution, svc service.Service, ctx context.Context) AuthResponse {
	return AuthResponse{
		TokenResponse: TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    3600, // 1 hour
		},
		User:        NewUserResponse(user, svc, ctx),
		Institution: NewInstitutionResponse(institution),
	}
}