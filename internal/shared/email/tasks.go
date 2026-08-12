package email

import "encoding/json"

// ================================================
// TASK TYPES
// ================================================

const (
	TypeVerificationOTP      = "email:verification_otp"
	TypeWelcomeIndividual    = "email:welcome_individual"
	TypeWelcomeInstitution   = "email:welcome_institution"
	TypeTwoFactorOTP         = "email:two_factor_otp"
	TypePasswordResetOTP     = "email:password_reset_otp"
	TypePasswordResetConfirm = "email:password_reset_confirm"
	TypeLoginNotification    = "email:login_notification"
)

// ================================================
// VERIFICATION PURPOSE
// ================================================

type VerificationPurpose string

const (
	PurposeRegistration  VerificationPurpose = "registration"
	PurposeEmailChange   VerificationPurpose = "email_change"
	PurposePhoneChange   VerificationPurpose = "phone_change"
	PurposePasswordReset VerificationPurpose = "password_reset"
	PurposeTwoFactor     VerificationPurpose = "two_factor"
)

// ================================================
// TASK STRUCTS
// ================================================

// VerificationOTPTask - Generic verification OTP for multiple purposes
type VerificationOTPTask struct {
	To      string              `json:"to"`
	Name    string              `json:"name"`
	OTP     string              `json:"otp"`
	Expires string              `json:"expires"`
	Purpose VerificationPurpose `json:"purpose"`
	Meta    map[string]string   `json:"meta,omitempty"`
}

func (t VerificationOTPTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}

// WelcomeIndividualTask - Individual professional welcome
type WelcomeIndividualTask struct {
	To   string `json:"to"`
	Name string `json:"name"`
}

func (t WelcomeIndividualTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}

// WelcomeInstitutionTask - Institution welcome
type WelcomeInstitutionTask struct {
	To              string `json:"to"`
	AdminName       string `json:"admin_name"`
	InstitutionName string `json:"institution_name"`
}

func (t WelcomeInstitutionTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}

// TwoFactorOTPTask
type TwoFactorOTPTask struct {
	To      string `json:"to"`
	Name    string `json:"name"`
	OTP     string `json:"otp"`
	Expires string `json:"expires"`
}

func (t TwoFactorOTPTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}

// PasswordResetOTPTask
type PasswordResetOTPTask struct {
	To      string `json:"to"`
	Name    string `json:"name"`
	OTP     string `json:"otp"`
	Expires string `json:"expires"`
}

func (t PasswordResetOTPTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}

// PasswordResetConfirmTask
type PasswordResetConfirmTask struct {
	To   string `json:"to"`
	Name string `json:"name"`
}

func (t PasswordResetConfirmTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}

// LoginNotificationTask
type LoginNotificationTask struct {
	To        string `json:"to"`
	Name      string `json:"name"`
	Time      string `json:"time"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

func (t LoginNotificationTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}