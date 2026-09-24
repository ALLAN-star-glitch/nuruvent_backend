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
		RedirectURL:       p.RedirectURL,       
		FailureReason:     p.FailureReason,
		InitiatedAt:       p.InitiatedAt,
		ExpiresAt:         p.ExpiresAt,
		CompletedAt:       p.CompletedAt,
		FailedAt:          p.FailedAt,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}



// toOrderResponse converts a domain Order to its wire format.
func toOrderResponse(o *paymentdomain.Order) OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, OrderItemResponse{
			TicketTypeID: it.TicketTypeID,
			Quantity:     it.Quantity,
			UnitPrice:    it.UnitPrice,
			Discount:     it.Discount,
			LineTotal:    it.LineTotal,
		})
	}

	return OrderResponse{
		ID:             o.ID,
		RegistrationID: o.RegistrationID,
		UserID:         o.UserID,
		GuestEmail:     o.GuestEmail,
		Currency:       o.Currency,
		Subtotal:       o.Subtotal,
		DiscountTotal:  o.DiscountTotal,
		TotalAmount:    o.TotalAmount,
		Status:         string(o.Status),
		ExpiresAt:      o.ExpiresAt,
		PaidAt:         o.PaidAt,
		CancelledAt:    o.CancelledAt,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
		Items:          items,
	}
}