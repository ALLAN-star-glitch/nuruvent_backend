// internal/app/adapters/certificate/payment_charger.go

package certificate

import (
	"context"

	certdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	paymentService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/service"
)

// PaymentChargerAdapter bridges the payment module's Service to the
// certificate module's PaymentCharger port.
type PaymentChargerAdapter struct {
	svc paymentService.Service
}

// NewPaymentChargerAdapter constructs a PaymentChargerAdapter.
func NewPaymentChargerAdapter(svc paymentService.Service) *PaymentChargerAdapter {
	return &PaymentChargerAdapter{svc: svc}
}

// CreateCertificateBatchOrder implements certdomain.PaymentCharger.
func (a *PaymentChargerAdapter) CreateCertificateBatchOrder(
	ctx context.Context,
	req certdomain.CertificateBatchOrderRequest,
) (*certdomain.CertificateBatchOrderResult, error) {
	recipients := make([]paymentService.CertificateBatchRecipientInput, 0, len(req.Recipients))
	for _, r := range req.Recipients {
		recipients = append(recipients, paymentService.CertificateBatchRecipientInput{
			UserID:      r.UserID,
			Name:        r.Name,
			Email:       r.Email,
			CompletedAt: r.CompletedAt,
		})
	}

	order, err := a.svc.CreateCertificateBatchOrder(ctx, paymentService.CreateCertificateBatchOrderCommand{
		EventID:           req.EventID,
		AccountID:         req.AccountID,
		TemplateID:        req.TemplateID,
		PricePerCertMinor: req.PricePerCertMinor,
		EventTitle:        req.EventTitle,
		EventDate:         req.EventDate,
		EventDuration:     req.EventDuration,
		IssuerName:        req.IssuerName,
		Recipients:        recipients,
		Currency:          req.Currency,
	})
	if err != nil {
		return nil, err
	}

	return &certdomain.CertificateBatchOrderResult{
		OrderID:     order.ID,
		AmountMinor: order.TotalAmount,
		Currency:    order.Currency,
	}, nil
}

// compile-time assertion
var _ certdomain.PaymentCharger = (*PaymentChargerAdapter)(nil)