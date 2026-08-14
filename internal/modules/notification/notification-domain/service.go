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
// INBOUND PORT: NotificationService Interface
// ============================================================

type NotificationService interface {
	// Verification OTP
	SendVerificationOTP(ctx context.Context, req SendOTPRequest) error

	// Welcome emails
	SendIndividualWelcome(ctx context.Context, req SendWelcomeRequest) error
	SendInstitutionWelcome(ctx context.Context, req SendInstitutionWelcomeRequest) error
	SendWelcome(ctx context.Context, req SendWelcomeRequest) error

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

type SendOTPRequest struct {
	To      string
	Name    string
	OTP     string
	Expires string
	Purpose VerificationPurpose
	Meta    map[string]string
}

type SendWelcomeRequest struct {
	To   string
	Name string
}

type SendInstitutionWelcomeRequest struct {
	To              string
	AdminName       string
	InstitutionName string
}

type SendTwoFactorRequest struct {
	To        string
	Name      string
	OTP       string
	Expires   string
	IPAddress string
	UserAgent string
}

type SendPasswordResetConfirmRequest struct {
	To   string
	Name string
}

type SendLoginNotificationRequest struct {
	To        string
	Name      string
	Time      string
	IPAddress string
	UserAgent string
}