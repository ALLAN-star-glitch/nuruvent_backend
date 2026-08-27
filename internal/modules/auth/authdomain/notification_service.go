// internal/modules/auth/authdomain/notification_service.go

package authdomain

import "context"

// ============================================================
// OUTBOUND PORT: NotificationService
// Defines what the auth module NEEDS for sending notifications
// ============================================================

// NotificationService defines the notification operations that auth module requires
type NotificationService interface {
	// ============================================================
	// UNIFIED OTP - Use this for all OTP purposes
	// ============================================================
	
	// SendOTP sends a verification OTP for any purpose
	// Purpose can be: "registration", "two_factor", "password_reset", "email_change", "phone_change"
	SendOTP(ctx context.Context, req SendOTPRequest) error

	// ============================================================
	// WELCOME EMAILS
	// ============================================================
	
	SendIndividualWelcome(ctx context.Context, req SendWelcomeRequest) error
	SendInstitutionWelcome(ctx context.Context, req SendInstitutionWelcomeRequest) error
	
	// SendInstitutionKYCWelcome sends KYC welcome email for institutions
	// This prompts institutions to complete their KYC verification within 7 days
	SendInstitutionKYCWelcome(ctx context.Context, req SendInstitutionKYCWelcomeRequest) error

	// ============================================================
	// PASSWORD RESET CONFIRM
	// ============================================================
	
	SendPasswordResetConfirm(ctx context.Context, req SendPasswordResetConfirmRequest) error

	// ============================================================
	// SECURITY NOTIFICATIONS
	// ============================================================
	
	SendLoginNotification(ctx context.Context, req SendLoginNotificationRequest) error


	SendNewInstitutionAccountNotification(ctx context.Context, req SendNewInstitutionAccountRegistrationRequest) error

	SendNewPersonalAccountNotification(ctx context.Context, req SendNewPersonalAccountRegistrationRequest) error
}

// ============================================================
// COMMANDS
// ============================================================

// SendOTPRequest - Unified OTP request for all purposes
type SendOTPRequest struct {
	To      string            // email or phone
	Name    string            // recipient name
	OTP     string            // generated OTP
	Expires string            // human-readable expiry (e.g., "5 minutes")
	Purpose string            // "registration", "two_factor", "password_reset", "email_change", "phone_change"
	Meta    map[string]string // additional context (IP, user agent, new_email, new_phone, etc.)
}

// SendWelcomeRequest - Individual welcome
type SendWelcomeRequest struct {
	To   string
	Name string
}

// SendInstitutionWelcomeRequest - Institution welcome
type SendInstitutionWelcomeRequest struct {
	To              string
	AdminName       string
	InstitutionName string
}

// ✅ NEW: SendInstitutionKYCWelcomeRequest - Institution KYC welcome
// Used to send KYC verification prompt to newly registered institutions
type SendInstitutionKYCWelcomeRequest struct {
	To              string // Admin email
	AdminName       string // Admin name
	InstitutionName string // Institution name
	InstitutionType string // Institution type (e.g., "university", "college", "training_center")
}

// To send the nuruvent admin an email notification of a new account registration
type SendNewInstitutionAccountRegistrationRequest struct {
	To			string
	NewAccountAdminName		string
	InstitutionName			string
	InstitutionType			string
}

type SendNewPersonalAccountRegistrationRequest struct {
	To  string
	NewAccountAdminName	string
}

// SendPasswordResetConfirmRequest - Password reset confirmation
type SendPasswordResetConfirmRequest struct {
	To   string
	Name string
}

// SendLoginNotificationRequest - Login notification
type SendLoginNotificationRequest struct {
	To        string
	Name      string
	Time      string
	IPAddress string
	UserAgent string
}