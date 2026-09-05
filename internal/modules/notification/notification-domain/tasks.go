// internal/modules/notification/notification-domain/tasks.go

package notificationdomain

// ============================================================
// TASK TYPES (Queue task names with "notification:" prefix)
// ============================================================

const (
	TaskVerificationOTP      = "notification:verification_otp"
	TaskWelcomeIndividual    = "notification:welcome_individual"
	TaskWelcomeInstitution   = "notification:welcome_institution"
	TaskPasswordResetConfirm = "notification:password_reset_confirm"
	TaskLoginNotification    = "notification:login_notification"
	TaskWelcomeInstitutionKYC     = "notification:welcome_institution_kyc" 
	TaskNewInstitutionAccountRegistration    = "notification:new_account_institution_registration_notice"
	TaskNewPersonalAccountRegistration = "notification:new_account_personal_registration_notice"
	// TEAM INVITATION TASKS
	TaskTeamInvite           = "notification:team_invite"
	TaskTeamInviteRegistration = "notification:team_invite_registration"
	TaskTeamInviteAccepted   = "notification:team_invite_accepted"
	TaskTeamInviteDeclined   = "notification:team_invite_declined"
)

// ============================================================
// TASK DATA STRUCTURES (Pure domain data, no JSON tags)
// ============================================================

// VerificationOTPTask - Unified OTP task for all purposes
// Purpose determines the type: registration, two_factor, password_reset, email_change, phone_change
type VerificationOTPTask struct {
	To      string              // email or phone
	Name    string              // recipient name
	OTP     string              // generated OTP
	Expires string              // human-readable expiry (e.g., "5 minutes")
	Purpose VerificationPurpose // EXPLICIT PURPOSE - registration, two_factor, password_reset, etc.
	Meta    map[string]string   // additional context (IP, user agent, new_email, new_phone, etc.)
}

// WelcomeIndividualTask - Individual welcome task
type WelcomeIndividualTask struct {
	To   string
	Name string
}

// WelcomeInstitutionTask - Institution welcome task
type WelcomeInstitutionTask struct {
	To              string
	AdminName       string
	InstitutionName string
	InstitutionEmail string
}

// PasswordResetConfirmTask - Password reset confirmation task
type PasswordResetConfirmTask struct {
	To   string
	Name string
}

// LoginNotificationTask - Login notification task
type LoginNotificationTask struct {
	To        string
	Name      string
	Time      string
	IPAddress string
	UserAgent string
}

type WelcomeInstitutionKYCTask struct {
	To              string
	AdminName       string
	InstitutionName string
	InstitutionType string
}

type NewInstitutionAccountRegistrationNotice struct {
	To						string
	NewAccountAdminName		string 
	InstitutionName			string
	InstitutionType		    string
}

type NewPersonalAccountRegistrationTask struct {
	To string
	NewAccountAdminName		string
}

// ============================================================
// ✅ TEAM INVITATION TASK STRUCTS (UPDATED - NO OTP)
// ============================================================

// TeamInviteTask - Team invitation email task for existing users
// Sent when: User exists → Direct add, just notify them
type TeamInviteTask struct {
	To         string // User's email
	UserName   string // User's name (if known)
	InvitedBy  string // Name of person who invited them
	TeamName   string // Name of the team (personal or institution)
	TeamID     string // ID of the team
	Role       string // Role: account_admin, event_manager, team_member
	InviteLink string // Link to accept the invitation
	ExpiresIn  string // Human-readable expiry (e.g., "7 days", or "N/A" for existing users)
}

// TeamInviteRegistrationTask - Invitation task for new users (with registration link)
// Sent when: User doesn't exist → Create invitation, send registration link
// ❌ NO OTP - User clicks the link from their email, which verifies the email
type TeamInviteRegistrationTask struct {
	To               string // User's email
	Name             string // User's name (optional, will be set during registration)
	Role             string // Role: account_admin, event_manager, team_member	
	TeamName         string // Name of the team
	InvitedBy        string // Name of person who invited them
	RegistrationLink string // Link to registration page with token
	ExpiresIn        string // Human-readable expiry (e.g., "7 days")
}

// TeamInviteAcceptedTask - Notification when invitation is accepted
type TeamInviteAcceptedTask struct {
	To        string // Admin's email
	AdminName string // Admin's name
	UserName  string // Name of user who accepted
	UserEmail string // Email of user who accepted
	TeamName  string // Name of the team
}

// TeamInviteDeclinedTask - Notification when invitation is declined
type TeamInviteDeclinedTask struct {
	To        string // Admin's email
	AdminName string // Admin's name
	UserName  string // Name of user who declined
	UserEmail string // Email of user who declined
	TeamName  string // Name of the team
}