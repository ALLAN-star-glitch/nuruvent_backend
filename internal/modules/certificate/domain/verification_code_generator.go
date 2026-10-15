// internal/modules/certificate/domain/verification_code_generator.go

package domain

// VerificationCodeGenerator produces cryptographically random codes.
//
// Implementations MUST guarantee >= MinVerificationCodeLength chars
// drawn from a URL-safe alphabet (BRD NFR-SEC-001).
type VerificationCodeGenerator interface {
	NewCode() (VerificationCode, error)
}