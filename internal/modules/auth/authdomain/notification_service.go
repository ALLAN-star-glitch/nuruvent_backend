package authdomain

import "context"

// ============================================================
// OUTBOUND PORT: NotificationService
// Defines what the auth module NEEDS for sending notifications
// ============================================================

// NotificationService defines the notification operations that auth module requires
type NotificationService interface {
	// Verification OTP
	SendVerificationOTP(ctx context.Context, req SendOTPRequest) error

	// Welcome emails
	SendIndividualWelcome(ctx context.Context, req SendWelcomeRequest) error
	SendInstitutionWelcome(ctx context.Context, req SendInstitutionWelcomeRequest) error

	// Two Factor Authentication
	SendTwoFactorOTP(ctx context.Context, req SendTwoFactorRequest) error

	// Password Reset
	SendPasswordResetOTP(ctx context.Context, req SendOTPRequest) error
	SendPasswordResetConfirm(ctx context.Context, req SendPasswordResetConfirmRequest) error

	// Security Notifications
	SendLoginNotification(ctx context.Context, req SendLoginNotificationRequest) error
}

// ============================================================
// COMMANDS
// ============================================================

// SendOTPRequest - Generic OTP request
type SendOTPRequest struct {
	To      string
	Name    string
	OTP     string
	Expires string
	Purpose string
	Meta    map[string]string
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

// SendTwoFactorRequest - 2FA OTP
type SendTwoFactorRequest struct {
	To        string
	Name      string
	OTP       string
	Expires   string
	IPAddress string
	UserAgent string
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