// internal/modules/payment/delivery/http/mappers.go

package http

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// toPaymentResponse converts a domain Payment to its wire format.
func toPaymentResponse(p *paymentdomain.Payment) PaymentResponse {
	return PaymentResponse{
		ID:                p.ID,
		OrderID:           p.OrderID,
		Provider:          p.Provider,
		Method:            string(p.Method),
		Amount:            p.Amount,
		Currency:          p.Currency,
		Status:            string(p.Status),
		ProviderReference: p.ProviderReference,
		FailureReason:     p.FailureReason,
		InitiatedAt:       p.InitiatedAt,
		ExpiresAt:         p.ExpiresAt,
		CompletedAt:       p.CompletedAt,
		FailedAt:          p.FailedAt,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}