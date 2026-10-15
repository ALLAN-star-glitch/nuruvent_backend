// internal/modules/payment/paymentdomain/certificate_batch_issuer.go

package paymentdomain

import "context"

// ============================================================
// OUTBOUND PORT: CertificateBatchIssuer
// ============================================================
//
// Invoked by the payment module when a certificate_batch order is
// paid. Implemented by a cross-module adapter that delegates to the
// certificate service.
//
// This is the organizer-pays counterpart to CertificateIssuer (which
// remains for the legacy attendee-pays flow).
type CertificateBatchIssuer interface {
	// IssueBatch issues all certificates in the order's batch context.
	// Idempotent — the certificate module returns existing
	// certificates for completions that already have one.
	IssueBatch(ctx context.Context, order *Order, payment *Payment) error
}