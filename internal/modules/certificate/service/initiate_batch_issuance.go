// internal/modules/certificate/service/initiate_batch_issuance.go

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// InitiateBatchIssuanceInput is the caller-supplied input.
type InitiateBatchIssuanceInput struct {
	EventID       string
	AccountID     string
	TemplateID    string
	EventTitle    string
	EventDate     time.Time
	EventDuration string
	IssuerName    string
	Recipients    []domain.BatchRecipient
}

// InitiateBatchIssuanceResult is what the caller gets back.
type InitiateBatchIssuanceResult struct {
	Type string // "free" | "paid"

	// Free path
	Issued int
	Reused int
	Failed int

	// Paid path
	OrderID       string
	AmountMinor   int64
	Currency      string
	AttendeeCount int
}

// InitiateBatchIssuance starts a certificate batch for an event.
func (s *serviceImpl) InitiateBatchIssuance(
	ctx context.Context,
	userID string,
	in InitiateBatchIssuanceInput,
) (*InitiateBatchIssuanceResult, error) {
	// ------------------------------------------------------------
	// 1. Validate input.
	// ------------------------------------------------------------
	if in.EventID == "" {
		return nil, domain.ErrContextIDRequired
	}
	if in.AccountID == "" {
		return nil, domain.ErrAccountIDRequired
	}
	if len(in.Recipients) == 0 {
		return nil, fmt.Errorf("at least one recipient is required")
	}

	// ------------------------------------------------------------
	// 2. Resolve the template + price.
	// ------------------------------------------------------------
	tmpl, err := s.resolveTemplateForBatch(ctx, in.AccountID, in.TemplateID)
	if err != nil {
		return nil, err
	}

	pricePerCert := int64(0)
	if tmpl.PriceMinor != nil {
		pricePerCert = *tmpl.PriceMinor
	}

	// ------------------------------------------------------------
	// 3. Free path — issue immediately.
	// ------------------------------------------------------------
	if pricePerCert == 0 {
		result, err := s.IssueBatch(ctx, domain.BatchIssuanceRequest{
			EventID:          in.EventID,
			AccountID:        in.AccountID,
			TemplateID:       tmpl.ID,
			EventTitle:       in.EventTitle,
			EventDate:        in.EventDate,
			EventDuration:    in.EventDuration,
			IssuerName:       in.IssuerName,
			Recipients:       in.Recipients,
			PaymentReference: "", // free
		})
		if err != nil {
			return nil, err
		}
		return &InitiateBatchIssuanceResult{
			Type:          "free",
			Issued:        result.Issued,
			Reused:        result.Reused,
			Failed:        result.Failed,
			AttendeeCount: len(in.Recipients),
		}, nil
	}

	// ------------------------------------------------------------
	// 4. Paid path — create a pending batch order.
	// ------------------------------------------------------------
	order, err := s.deps.Payment.CreateCertificateBatchOrder(ctx, domain.CertificateBatchOrderRequest{
		EventID:           in.EventID,
		AccountID:         in.AccountID,
		TemplateID:        tmpl.ID,
		PricePerCertMinor: pricePerCert,
		EventTitle:        in.EventTitle,
		EventDate:         in.EventDate,
		EventDuration:     in.EventDuration,
		IssuerName:        in.IssuerName,
		Recipients:        in.Recipients,
		Currency:          domain.DefaultCurrency,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrPaymentChargeFailed, err)
	}

	return &InitiateBatchIssuanceResult{
		Type:          "paid",
		OrderID:       order.OrderID,
		AmountMinor:   order.AmountMinor,
		Currency:      order.Currency,
		AttendeeCount: len(in.Recipients),
	}, nil
}

// resolveTemplateForBatch picks the template for a batch, in order:
// explicit templateID → account's default → system default.
func (s *serviceImpl) resolveTemplateForBatch(
	ctx context.Context,
	accountID, templateID string,
) (*domain.Template, error) {
	if templateID != "" {
		t, err := s.deps.Templates.FindByID(ctx, templateID)
		if err != nil {
			return nil, err
		}
		// System templates are accessible to any account.
		// Otherwise the template must belong to this account.
		if !t.IsSystem {
			if t.AccountID == nil || *t.AccountID != accountID {
				return nil, domain.ErrTemplateNotFound
			}
		}
		return t, nil
	}

	if t, err := s.deps.Templates.FindDefaultForAccount(ctx, accountID); err == nil {
		return t, nil
	}

	return s.deps.Templates.FindSystemBySlug(ctx, domain.SystemTemplateDefault)
}