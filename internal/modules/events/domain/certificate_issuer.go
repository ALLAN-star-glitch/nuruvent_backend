// internal/modules/events/domain/certificate_issuer.go

package domain

import (
	"context"
	"time"
)

// ============================================================
// OUTBOUND PORT: CertificateIssuer
// ============================================================
//
// Issues certificates for an event's qualified attendees. Implemented
// by a cross-module adapter that delegates to the certificate module.
//
// See internal/app/adapters/events/certificate_adapter.go.
//
// The events module owns the qualification decision (who attended and
// qualifies for a certificate). The certificate module owns issuance,
// verification, delivery, and templates. This port carries the minimum
// data the certificate module needs — no reference to event internals.
type CertificateIssuer interface {
	// IssueForEvent issues certificates for every recipient in the
	// request. Idempotent per recipient — callers can retry safely;
	// already-issued certificates are returned, not duplicated.
	//
	// Returns per-recipient results. A single failure does not abort
	// the batch.
	IssueForEvent(
		ctx context.Context,
		req IssueCertificatesRequest,
	) (*IssueCertificatesResult, error)
}

// IssueCertificatesRequest carries everything the certificate module
// needs to issue certificates for one event.
type IssueCertificatesRequest struct {
	// EventID is the events module's own identifier for the event.
	// The certificate module treats this as an opaque context_id.
	EventID string

	// AccountID is the account that owns the event. Certificates are
	// issued under this account.
	AccountID string

	// TemplateID is optional. Empty means use the account's default,
	// or a system template if none is set.
	TemplateID string

	// EventTitle is rendered into the certificate.
	EventTitle string

	// EventDate is the date the event occurred, rendered into the
	// certificate.
	EventDate time.Time

	// EventDuration is a human-readable duration string ("2 hours"),
	// rendered into the certificate. Optional.
	EventDuration string

	// IssuerName is the organization name printed on the certificate.
	IssuerName string

	// Recipients are the attendees who qualify for a certificate.
	Recipients []CertificateRecipient
}

// CertificateRecipient is one attendee who qualifies.
type CertificateRecipient struct {
	// UserID is the Nuruvent user ID, if the recipient has an account.
	// Empty for guest recipients.
	UserID string

	// Name and Email are snapshotted onto the certificate.
	Name  string
	Email string

	// CompletedAt is when the recipient completed the event. Usually
	// the last session's end time.
	CompletedAt time.Time
}

// IssueCertificatesResult summarizes the outcome.
type IssueCertificatesResult struct {
	Issued int
	Reused int
	Failed int
	Errors []string
}