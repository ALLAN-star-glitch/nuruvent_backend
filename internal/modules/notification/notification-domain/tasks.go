// internal/modules/notification/notification-domain/tasks.go

package notificationdomain

import types "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types/emails"

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
	
	// ============================================================
	// TEAM INVITATION TASKS (UPDATED)
	// ============================================================
	TaskTeamInviteExistingUser  = "notification:team_invite_existing_user"  // ✅ Renamed - for existing users
	TaskTeamInviteRegistration  = "notification:team_invite_registration"   // ✅ New users (no OTP)
	TaskTeamInviteAccepted      = "notification:team_invite_accepted"       // ✅ Admin notification
	TaskTeamInviteDeclined      = "notification:team_invite_declined"       // ✅ Admin notification
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
// ✅ TEAM INVITATION TASK STRUCTS (UPDATED - NO ROLE, NO OTP, AI-READY)
// ============================================================

// TeamInviteExistingUserTask - Invitation for existing users (has account)
// Sent when: User already has a Nuruvent account
// ✅ NO ROLE - Roles are inherited from account level
// ✅ NO OTP - User clicks accept link to join
// ✅ AI-READY - PersonalizedContent field for future AI integration
type TeamInviteExistingUserTask struct {
	To         string // User's email
	UserName   string // User's name
	InvitedBy  string // Name of person who invited them
	TeamName   string // Name of the team
	TeamID     string // ID of the team
	AcceptLink string // One-click accept link: https://nuruvent.com/invitations/accept?token=xxx
	ExpiresIn  string // Human-readable expiry (e.g., "7 days")
	
	// ✅ AI-Ready: Personalized content (optional, for future use)
	PersonalizedContent *types.PersonalizedInvitationContent
}

// TeamInviteRegistrationTask - Invitation for new users (no account yet)
// Sent when: User does NOT have a Nuruvent account
// ✅ NO ROLE - Roles are inherited from account level
// ✅ NO OTP - User clicks registration link with token embedded
// ✅ AI-READY - PersonalizedContent field for future AI integration
type TeamInviteRegistrationTask struct {
	To               string // User's email
	Name             string // User's name (optional, will be set during registration)
	InvitedBy        string // Name of person who invited them
	TeamName         string // Name of the team
	TeamID           string // ID of the team
	RegistrationLink string // Registration link with token: https://nuruvent.com/register?token=xxx
	ExpiresIn        string // Human-readable expiry (e.g., "7 days")
	
	// ✅ AI-Ready: Personalized content (optional, for future use)
	PersonalizedContent *types.PersonalizedInvitationContent 
}

// TeamInviteAcceptedTask - Notification to admin when invitation is accepted
// ✅ AI-READY - PersonalizedContent field for future AI integration
type TeamInviteAcceptedTask struct {
	To        string // Admin's email
	AdminName string // Admin's name
	UserName  string // Name of user who accepted
	UserEmail string // Email of user who accepted
	TeamName  string // Name of the team
	TeamID    string // ID of the team
	
	// ✅ AI-Ready: Personalized content (optional, for future use)
	PersonalizedContent *types.PersonalizedInvitationContent 
}

// TeamInviteDeclinedTask - Notification to admin when invitation is declined
// ✅ AI-READY - PersonalizedContent field for future AI integration
type TeamInviteDeclinedTask struct {
	To        string // Admin's email
	AdminName string // Admin's name
	UserName  string // Name of user who declined
	UserEmail string // Email of user who declined
	TeamName  string // Name of the team
	TeamID    string // ID of the team
	
	// ✅ AI-Ready: Personalized content (optional, for future use)
	PersonalizedContent *types.PersonalizedInvitationContent 
}



