package authdomain



// ============================================================
// OUTBOUND PORT: EmailService
// ============================================================

type EmailService interface {
    SendVerificationOTP(to, name, otp, expiry, purpose string, metadata map[string]string) error
    SendWelcome(to, name string) error
    SendPasswordResetOTP(to, name, otp string) error
    SendPasswordResetConfirm(to, name string) error
    SendLoginNotification(to, name, ipAddress, userAgent string) error
}