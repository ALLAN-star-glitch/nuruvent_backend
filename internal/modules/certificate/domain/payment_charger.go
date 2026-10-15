// internal/modules/certificate/domain/payment_charger.go

package domain

import (
	"context"
	"time"
)

// ============================================================
// OUTBOUND PORT: PaymentCharger
// ============================================================
//
// Creates a payment order for a certificate batch. Implemented by a
// cross-module adapter that delegates to the payment module.
//
// The certificate module never talks to the payment module directly.
// This port is the only seam.
type PaymentCharger interface {
	// CreateCertificateBatchOrder creates a pending payment order for
	// a batch of certificates. Idempotent per event — a repeat call
	// returns the existing pending order.
	CreateCertificateBatchOrder(
		ctx context.Context,
		req CertificateBatchOrderRequest,
	) (*CertificateBatchOrderResult, error)
}

// CertificateBatchOrderRequest carries everything the payment module
// needs to create a pending batch order.
type CertificateBatchOrderRequest struct {
	EventID           string
	AccountID         string
	TemplateID        string
	PricePerCertMinor int64

	EventTitle    string
	EventDate     time.Time
	EventDuration string
	IssuerName    string

	Recipients []BatchRecipient

	Currency string
}

// CertificateBatchOrderResult is what the payment module returns
// after creating the pending order.
type CertificateBatchOrderResult struct {
	OrderID     string
	AmountMinor int64
	Currency    string
}