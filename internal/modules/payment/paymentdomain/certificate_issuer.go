// internal/modules/payment/paymentdomain/certificate_issuer.go

package paymentdomain

import "context"

// ============================================================
// OUTBOUND PORT: CertificateIssuer
// ============================================================
//
// Invoked by the payment module when a certificate order is paid.
// Implemented by a cross-module adapter that delegates to the
// certificate service.
//
// This is the second half of the paid-certificate flow:
//
//  1. The certificate module calls payment.CreateCertificateOrder to
//     create a pending order.
//  2. The attendee pays via the provider.
//  3. The provider webhook fires.
//  4. ConfirmPayment marks the order paid.
//  5. THIS PORT is called to issue the certificate.
//
// See internal/app/adapters/payment/certificate_issuer.go.
type CertificateIssuer interface {
	// IssueFromPaidOrder issues the certificate described by the
	// order's certificate context. Idempotent — the certificate
	// module returns an existing certificate if one already exists
	// for the completion.
	IssueFromPaidOrder(
		ctx context.Context,
		order *Order,
		payment *Payment,
	) error
}