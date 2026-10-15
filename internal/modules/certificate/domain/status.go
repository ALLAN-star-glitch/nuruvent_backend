// internal/modules/certificate/domain/status.go

package domain

// Status is the lifecycle state of a certificate.
type Status string

const (
	StatusActive  Status = "active"
	StatusRevoked Status = "revoked"
)

// IsValid reports whether s is a known status.
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusRevoked:
		return true
	default:
		return false
	}
}

// CanVerify reports whether a certificate in this status can pass
// public verification.
func (s Status) CanVerify() bool { return s == StatusActive }

// CanRevoke reports whether a certificate in this status can be revoked.
func (s Status) CanRevoke() bool { return s == StatusActive }

// String returns the raw status.
func (s Status) String() string { return string(s) }