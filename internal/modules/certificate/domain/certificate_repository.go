// internal/modules/certificate/domain/certificate_repository.go

package domain

import (
	"context"
	"time"
)

// CertificateRepository persists certificates and their deliveries.
//
// Implementations must enforce uniqueness on
// (context_type, context_id, recipient_user_id/recipient_email) so the
// application service can rely on idempotency (BRD FR-CI-003/005).
type CertificateRepository interface {
	// Create inserts a new certificate. Returns ErrAlreadyIssued if a
	// certificate already exists for the same completion identity.
	Create(ctx context.Context, c *Certificate) error

	// FindByID loads a certificate by its primary key.
	FindByID(ctx context.Context, id string) (*Certificate, error)

	// FindByVerificationCode loads a certificate by its public code.
	FindByVerificationCode(ctx context.Context, code VerificationCode) (*Certificate, error)

	// FindByIdentity returns the existing certificate for a completion
	// identity, or ErrCertificateNotFound if none exists.
	FindByIdentity(ctx context.Context, id CompletionIdentity) (*Certificate, error)

	// Update persists changes to an existing certificate.
	Update(ctx context.Context, c *Certificate) error

	// ListByAccount returns certificates issued by an account.
	ListByAccount(ctx context.Context, accountID string, f CertificateFilter) ([]*Certificate, error)

	// ListByRecipient returns certificates for a recipient user.
	ListByRecipient(ctx context.Context, userID string, f CertificateFilter) ([]*Certificate, error)

	// RecordDelivery persists a delivery attempt.
	RecordDelivery(ctx context.Context, d *Delivery) error
}

// CertificateFilter narrows certificate listings.
type CertificateFilter struct {
	ContextType string
	ContextID   string
	From        *time.Time
	To          *time.Time
	Status      *Status
	Limit       int
	Offset      int
}