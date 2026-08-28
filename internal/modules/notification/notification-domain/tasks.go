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