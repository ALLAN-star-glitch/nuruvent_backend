// internal/modules/payment/paymentdomain/certificate_context.go

package paymentdomain

import "time"

// CertificateContext carries the input a certificate order needs so
// the payment module can hand it to the certificate module on
// successful payment.
//
// The payment module treats the fields opaquely — it stores them,
// forwards them, and never interprets what "context_type" or
// "context_id" mean.
//
// Persistence (JSONB on the orders row) is handled by the GORM
// adapter, not the domain. This type is pure.
type CertificateContext struct {
	// ContextType is the calling module's identifier for the kind of
	// context ("event", "course", ...). Opaque to this module.
	ContextType string

	// ContextID is the caller's own identifier for the context.
	ContextID string

	// TemplateID is optional. Empty means use the account's default,
	// or a system template.
	TemplateID string

	// Recipient identity
	RecipientName   string
	RecipientEmail  string
	RecipientUserID string

	// CompletedAt is when the completion happened. Distinct from the
	// issue time.
	CompletedAt time.Time

	// AccountID is the account the certificate belongs to (the
	// issuer). Also the BilledAccountID on the order.
	AccountID string
}