package notificationdomain

// ============================================================
// TASK TYPES (Queue task names with "notification:" prefix)
// ============================================================

const (
    TaskVerificationOTP      = "notification:verification_otp"
    TaskWelcomeIndividual    = "notification:welcome_individual"
    TaskWelcomeInstitution   = "notification:welcome_institution"
    TaskTwoFactorOTP         = "notification:two_factor_otp"
    TaskPasswordResetOTP     = "notification:password_reset_otp"
    TaskPasswordResetConfirm = "notification:password_reset_confirm"
    TaskLoginNotification    = "notification:login_notification"
)

// ============================================================
// TASK DATA STRUCTURES (Pure domain data, no JSON tags)
// ============================================================

// VerificationOTPTask - Generic verification OTP task
type VerificationOTPTask struct {
    To      string
    Name    string
    OTP     string
    Expires string
    Purpose VerificationPurpose
    Meta    map[string]string
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
}

// TwoFactorOTPTask - 2FA OTP task
type TwoFactorOTPTask struct {
    To        string
    Name      string
    OTP       string
    Expires   string
    IPAddress string
    UserAgent string
}

// PasswordResetOTPTask - Password reset OTP task
type PasswordResetOTPTask struct {
    To      string
    Name    string
    OTP     string
    Expires string
}

// PasswordResetConfirmTask - Password reset confirm task
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