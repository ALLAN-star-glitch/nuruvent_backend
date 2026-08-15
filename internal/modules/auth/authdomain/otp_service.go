package authdomain


// ============================================================
// OUTBOUND PORT: OTPService
// Defines what the authdomain
//  NEEDS for OTP operations
// ============================================================

type OTPService interface {
    GenerateOTP() string
    StoreOTP(email, otp string) error
    GetOTP(email string) (string, error)
    DeleteOTP(email string) error
    StoreTwoFactorOTP(email, otp string) error
    GetTwoFactorOTP(email string) (string, error)
    DeleteTwoFactorOTP(email string) error
    StoreUserData(email string, data map[string]any) error
    GetUserData(email string) (map[string]string, error)
    DeleteUserData(email string) error
    StoreResetData(email, otp, newPassword string) error
    GetResetData(email string) (map[string]string, error)
    DeleteResetData(email string) error
    SendOTPEmail(to, name, otp string) error
    SendTwoFactorOTP(to, name, otp, ipAddress, userAgent string) error
    SendPasswordResetOTP(to, name, otp string) error
}