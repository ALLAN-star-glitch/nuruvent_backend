// internal/modules/payment/paymentdomain/batch_context.go

package paymentdomain

import "time"

// BatchContext carries everything the payment module needs to know
// about a certificate batch order so it can hand the batch to the
// certificate module on successful payment.
//
// The payment module stores it opaquely and forwards it back. It
// never interprets event_id, template_id, or any other field.
type BatchContext struct {
	// Event the batch belongs to.
	EventID   string
	AccountID string

	// The certificate template + the price snapshot at batch time.
	TemplateID       string
	PricePerCertMinor int64

	// Denormalized event metadata for the certificate render.
	EventTitle    string
	EventDate     time.Time
	EventDuration string
	IssuerName    string

	// AttendeeCount is the number of certificates the batch covers.
	AttendeeCount int

	// Recipients is the snapshot of who qualifies at batch time.
	// Frozen here so issuance cannot drift if attendance changes
	// between order creation and payment success.
	Recipients []BatchRecipient
}

// BatchRecipient is one attendee in a batch.
type BatchRecipient struct {
	UserID      string
	Name        string
	Email       string
	CompletedAt time.Time
}