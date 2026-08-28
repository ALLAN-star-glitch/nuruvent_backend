// internal/modules/notification/notification-domain/service.go

package notificationdomain

import "context"

// ============================================================
// VERIFICATION PURPOSE
// ============================================================

type VerificationPurpose string

const (
	PurposeRegistration  VerificationPurpose = "registration"
	PurposeEmailChange   VerificationPurpose = "email_change"
	PurposePhoneChange   VerificationPurpose = "phone_change"
	PurposePasswordReset VerificationPurpose = "password_reset"
	PurposeTwoFactor     VerificationPurpose = "two_factor"
)

func (p VerificationPurpose) String() string {
	return string(p)
}

func (p VerificationPurpose) IsValid() bool {
	switch p {
	case PurposeRegistration, PurposeEmailChange, PurposePhoneChange, PurposePasswordReset, PurposeTwoFactor:
		return true
	default:
		return false
	}
}

// ============================================================
// OTP CONFIG - Different purposes have different requirements
// ============================================================

type OTPConfig struct {
	Length          int
	ExpirySeconds   int64
	MaxAttempts     int
	RateLimitMax    int
	RateLimitWindow int64 // seconds
}

var OTPConfigs = map[VerificationPurpose]OTPConfig{
	PurposeRegistration: {
		Length:          6,
		ExpirySeconds:   3600, // 1 hour
		MaxAttempts:     5,
		RateLimitMax:    5,
		RateLimitWindow: 300, // 5 minutes
	},
	PurposeTwoFactor: {
		Length:          6,
		ExpirySeconds:   300, // 5 minutes
		MaxAttempts:     3,
		RateLimitMax:    3,
		RateLimitWindow: 300, // 5 minutes
	},
	PurposePasswordReset: {
		Length:          6,
		ExpirySeconds:   900, // 15 minutes
		MaxAttempts:     3,
		RateLimitMax:    3,
		RateLimitWindow: 3600, // 1 hour (stricter)
	},
	PurposeEmailChange: {
		Length:          6,
		ExpirySeconds:   3600, // 1 hour
		MaxAttempts:     3,
		RateLimitMax:    3,
		RateLimitWindow: 600, // 10 minutes
	},
	PurposePhoneChange: {
		Length:          6,
		ExpirySeconds:   600, // 10 minutes
		MaxAttempts:     3,
		RateLimitMax:    3,
		RateLimitWindow: 600, // 10 minutes
	},
}

// ============================================================
// INBOUND PORT: NotificationService Interface
// ============================================================

type NotificationService interface {
	// ============================================================
	// VERIFICATION OTP - Unified method with explicit purpose
	// ============================================================
	
	// SendOTP sends a verification OTP for any purpose.
	// The purpose determines the OTP configuration (length, expiry, etc.)
	SendOTP(ctx context.Context, req SendOTPRequest) error
	
	// VerifyOTP verifies an OTP for a specific purpose.
	// This ensures OTPs cannot be reused across different purposes.
	VerifyOTP(ctx context.Context, req VerifyOTPRequest) error

	// ============================================================
	// WELCOME EMAILS
	// ============================================================
	
	SendIndividualWelcome(ctx context.Context, req SendWelcomeRequest) error
	SendInstitutionWelcome(ctx context.Context, req SendInstitutionWelcomeRequest) error
	SendInstitutionKYCWelcome(ctx context.Context, req SendInstitutionKYCWelcomeRequest) error
	SendNewInstitutionAccountNotification(ctx context.Context, req SendNewInstitutionAccountRegistrationRequest) error
	SendNewPersonalAccountNotification(ctx context.Context, req SendNewPersonalAccountRegistrationRequest) error

	// ============================================================
	// SECURITY NOTIFICATIONS
	// ============================================================
	
	SendLoginNotification(ctx context.Context, req SendLoginNotificationRequest) error
	SendPasswordResetConfirm(ctx context.Context, req SendPasswordResetConfirmRequest) error
}

// ============================================================
// DOMAIN COMMANDS (no JSON tags)
// ============================================================

// SendOTPRequest - Unified request for all OTP purposes
type SendOTPRequest struct {
	To      string              // email or phone
	Name    string              // recipient name
	OTP     string              // generated OTP
	Expires string              // human-readable expiry (e.g., "5 minutes")
	Purpose VerificationPurpose // EXPLICIT PURPOSE - CRITICAL!
	Meta    map[string]string   // additional context (IP, user agent, etc.)
}

// VerifyOTPRequest - Unified verification request
type VerifyOTPRequest struct {
	To      string              // email or phone
	OTP     string              // OTP to verify
	Purpose VerificationPurpose // Must match the purpose used when sending
	Meta    map[string]string   // additional context
}

// SendWelcomeRequest - Welcome email for individual users
type SendWelcomeRequest struct {
	To   string
	Name string
}

// SendInstitutionWelcomeRequest - Welcome email for institution admins
type SendInstitutionWelcomeRequest struct {
	To              string
	AdminName       string
	InstitutionName string
	InstitutionEmail string
}

// ✅ NEW: SendInstitutionKYCWelcomeRequest - Welcome email with KYC requirements
type SendInstitutionKYCWelcomeRequest struct {
	To              string
	AdminName       string
	InstitutionName string
	InstitutionType string
}

type SendNewInstitutionAccountRegistrationRequest struct {
	TO string
	NewAccountAdminName string
	InstitutionName  string
	InstitutionType string
}

type SendNewPersonalAccountRegistrationRequest struct {
	To string
	NewAccountAdminName string
}

// SendPasswordResetConfirmRequest - Password reset confirmation email
type SendPasswordResetConfirmRequest struct {
	To   string
	Name string
}

// SendLoginNotificationRequest - Login notification email
type SendLoginNotificationRequest struct {
	To        string
	Name      string
	Time      string
	IPAddress string
	UserAgent string
}