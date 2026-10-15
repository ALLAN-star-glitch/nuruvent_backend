// internal/modules/certificate/domain/batch_issuer.go

package domain

import "time"

// ============================================================
// INPUT TYPES: Certificate batch issuance
// ============================================================
//
// Called by the payment webhook adapter after a certificate_batch
// order is paid. Carries the frozen batch snapshot so issuance cannot
// drift from what the organizer paid for.

// BatchIssuanceRequest describes one batch of certificates to issue.
type BatchIssuanceRequest struct {
	EventID       string
	AccountID     string
	TemplateID    string
	EventTitle    string
	EventDate     time.Time
	EventDuration string
	IssuerName    string
	Recipients    []BatchRecipient

	// PaymentReference ties the issued certificates to the payment
	// that paid for them. Stored on each certificate row.
	PaymentReference string
}

// BatchRecipient is one attendee in a batch.
type BatchRecipient struct {
	UserID      string
	Name        string
	Email       string
	CompletedAt time.Time
}

// BatchIssuanceResult summarizes the batch issuance.
type BatchIssuanceResult struct {
	Issued int
	Reused int
	Failed int
	Errors []string
}