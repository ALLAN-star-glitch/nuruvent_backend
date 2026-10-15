// internal/modules/certificate/domain/certificate.go

package domain

import "time"

// Certificate is the aggregate root for an issued certificate.
//
// A certificate snapshots everything it needs to render: recipient
// name/email, template reference, metadata, completion date. It never
// consults the original context (event/course) at render time
// (BRD FR-CI-008).
type Certificate struct {
	ID               string
	AccountID        string
	TemplateID       string
	VerificationCode VerificationCode

	// Recipient snapshot
	RecipientUserID string // optional
	RecipientName   string
	RecipientEmail  string

	// Opaque context reference
	ContextType string
	ContextID   string

	// Snapshot of the metadata used at issue time.
	Metadata Metadata

	// CompletedAt is when the learning activity completed; IssuedAt is
	// when this certificate was created. They are distinct
	// (BRD FR-CI-011).
	CompletedAt time.Time
	IssuedAt    time.Time

	// PaymentReference is non-empty for paid certificates.
	PaymentReference string

	Status           Status
	RevokedAt        *time.Time
	RevocationReason string

	// Cached PDF location, populated after first render.
	PDFURL         string
	PDFGeneratedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsRevoked reports whether the certificate has been revoked.
func (c Certificate) IsRevoked() bool {
	return c.Status == StatusRevoked
}

// CanVerify reports whether the certificate passes public verification.
func (c Certificate) CanVerify() bool {
	return c.Status.CanVerify()
}

// Identity returns the completion identity this certificate was issued
// for. Used by the repository for idempotency lookups.
func (c Certificate) Identity() CompletionIdentity {
	return CompletionIdentity{
		ContextType: c.ContextType,
		ContextID:   c.ContextID,
		RecipientID: recipientKey(c.RecipientUserID, c.RecipientEmail),
	}
}

// Revoke transitions the certificate to revoked. The reason is
// optional but recorded.
func (c *Certificate) Revoke(reason string, at time.Time) error {
	if !c.Status.CanRevoke() {
		return ErrCertificateAlreadyRevoked
	}
	ts := at.UTC()
	c.Status = StatusRevoked
	c.RevokedAt = &ts
	c.RevocationReason = reason
	return nil
}

func recipientKey(userID, email string) string {
	if userID != "" {
		return "user:" + userID
	}
	return "email:" + email
}