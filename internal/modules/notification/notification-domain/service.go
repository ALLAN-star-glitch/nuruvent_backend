// internal/modules/notification/notification-domain/notification.go

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
	PurposeTeamInvite    VerificationPurpose = "team_invite"
)

func (p VerificationPurpose) String() string {
	return string(p)
}

func (p VerificationPurpose) IsValid() bool {
	switch p {
	case PurposeRegistration, PurposeEmailChange, PurposePhoneChange, PurposePasswordReset, PurposeTwoFactor, PurposeTeamInvite:
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
	PurposeTeamInvite: {
		Length:          6,
		ExpirySeconds:   86400, // 24 hours
		MaxAttempts:     5,
		RateLimitMax:    5,
		RateLimitWindow: 300, // 5 minutes
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

	// ============================================================
	// ✅ TEAM INVITATIONS
	// ============================================================
	
	// SendTeamInvite sends a team invitation email to an existing user
	// User is already registered, so no OTP needed, just notification
	SendTeamInvite(ctx context.Context, req SendTeamInviteRequest) error
	
	// SendTeamInviteRegistration sends invitation to a new user with registration link
	// User is NOT registered, so they get a registration link (OTP during registration)
	SendTeamInviteRegistration(ctx context.Context, req SendTeamInviteRegistrationRequest) error
	
	// SendTeamInviteAccepted sends notification to admin when user accepts invitation
	SendTeamInviteAccepted(ctx context.Context, req SendTeamInviteAcceptedRequest) error
	
	// SendTeamInviteDeclined sends notification to admin when user declines invitation
	SendTeamInviteDeclined(ctx context.Context, req SendTeamInviteDeclinedRequest) error
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

// ============================================================
// WELCOME EMAIL COMMANDS
// ============================================================

// SendWelcomeRequest - Welcome email for individual users
type SendWelcomeRequest struct {
	To   string
	Name string
}

// SendInstitutionWelcomeRequest - Welcome email for institution admins
type SendInstitutionWelcomeRequest struct {
	To               string
	AdminName        string
	InstitutionName  string
	InstitutionEmail string
}

// SendInstitutionKYCWelcomeRequest - Welcome email with KYC requirements
type SendInstitutionKYCWelcomeRequest struct {
	To              string
	AdminName       string
	InstitutionName string
	InstitutionType string
}

// SendNewInstitutionAccountRegistrationRequest - Notification to internal admin about new institution account
type SendNewInstitutionAccountRegistrationRequest struct {
	To                  string
	NewAccountAdminName string
	InstitutionName     string
	InstitutionType     string
}

// SendNewPersonalAccountRegistrationRequest - Notification to internal admin about new personal account
type SendNewPersonalAccountRegistrationRequest struct {
	To                  string
	NewAccountAdminName string
}

// ============================================================
// SECURITY NOTIFICATION COMMANDS
// ============================================================

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

// ============================================================
// ✅ TEAM INVITATION COMMANDS (Updated for Team Module)
// ============================================================

// SendTeamInviteRequest - Team invitation email for existing users
// Sent when: User already exists → Direct add, just notify them
type SendTeamInviteRequest struct {
	To          string // User's email
	UserName    string // User's name (if known)
	InvitedBy   string // Name of person who invited them
	TeamName    string // Name of the team (personal or institution)
	TeamID      string // ID of the team
	Role        string // Role: account_admin, event_manager, team_member
	InviteLink  string // Link to accept the invitation
	ExpiresIn   string // Human-readable expiry (e.g., "7 days", or "N/A" for existing users)
}

// SendTeamInviteRegistrationRequest - Invitation email for new users (with registration link)
// Sent when: User doesn't exist → Create invitation, send registration link
type SendTeamInviteRegistrationRequest struct {
	To               string // User's email
	Name             string // User's name (optional, will be set during registration)
	InvitedBy        string // Name of person who invited them
	TeamName         string // Name of the team (personal or institution)
	Role             string // Role: account_admin, event_manager, team_member
	RegistrationLink string // Link to registration page with token
	ExpiresIn        string // Human-readable expiry (e.g., "7 days")
}

// SendTeamInviteAcceptedRequest - Notification to admin when user accepts invitation
type SendTeamInviteAcceptedRequest struct {
	To         string // Admin's email
	AdminName  string // Admin's name
	UserName   string // Name of user who accepted
	UserEmail  string // Email of user who accepted
	TeamName   string // Name of the team
}

// SendTeamInviteDeclinedRequest - Notification to admin when user declines invitation
type SendTeamInviteDeclinedRequest struct {
	To         string // Admin's email
	AdminName  string // Admin's name
	UserName   string // Name of user who declined
	UserEmail  string // Email of user who declined
	TeamName   string // Name of the team
}