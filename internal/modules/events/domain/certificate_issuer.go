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
// The events module forwards certificate batch requests to the
// certificate module through this port. The certificate module
// decides whether the batch is free (issue immediately) or paid
// (create a pending order that must be settled before issuance).
//
// The events module owns the qualification decision (who attended
// and qualifies for a certificate). The certificate module owns
// issuance, verification, delivery, templates, and pricing. This
// port carries the minimum data the certificate module needs — no
// reference to event internals.
//
// See internal/app/adapters/events/certificate_issuer_holder.go.
type CertificateIssuer interface {
	// InitiateBatchIssuance starts a certificate batch for an event.
	//
	// Free batches issue immediately and return Type = "free" with
	// Issued/Reused/Failed counts.
	//
	// Paid batches create a pending payment order and return
	// Type = "paid" with the OrderID and amount. Issuance happens
	// later, when the payment webhook fires.
	//
	// Idempotent per event on both paths.
	InitiateBatchIssuance(
		ctx context.Context,
		req InitiateBatchIssuanceRequest,
	) (*InitiateBatchIssuanceResult, error)
}

// ============================================================
// REQUEST / RESULT TYPES
// ============================================================

// InitiateBatchIssuanceRequest carries everything the certificate
// module needs to start a batch.
type InitiateBatchIssuanceRequest struct {
	// UserID is the caller (for authz forwarding).
	UserID string

	// EventID is the events module's own identifier for the event.
	// The certificate module treats this as an opaque context_id.
	EventID string

	// AccountID is the account that owns the event. Certificates are
	// issued under this account and the organizer is billed.
	AccountID string

	// TemplateID is optional. Empty means use the account's default,
	// or a system template if none is set. The template's price
	// determines whether the batch is free or paid.
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

// InitiateBatchIssuanceResult is the outcome of a batch initiation.
//
// Exactly one set of fields is populated based on Type:
//   - Type = "free" → Issued, Reused, Failed
//   - Type = "paid" → OrderID, AmountMinor, Currency
type InitiateBatchIssuanceResult struct {
	// Type is either "free" or "paid".
	Type string

	// ---- Free path ----
	Issued int
	Reused int
	Failed int

	// ---- Paid path ----
	OrderID     string
	AmountMinor int64
	Currency    string

	// ---- Shared ----
	AttendeeCount int
}