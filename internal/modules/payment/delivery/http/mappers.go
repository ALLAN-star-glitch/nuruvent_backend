// internal/modules/payment/delivery/http/mappers.go

package http

import (
	"time"

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
		AccessCode:        p.AccessCode,
		FailedAt:          p.FailedAt,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}

// toOrderResponse converts a domain Order to its wire format.
//
// Order items are populated for registration orders. Batch context
// is populated for certificate_batch orders. Both are empty for the
// other kind.
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
		ID:                 o.ID,
		Kind:               string(o.Kind),
		RegistrationID:     o.RegistrationID,
		UserID:             o.UserID,
		GuestEmail:         o.GuestEmail,
		Currency:           o.Currency,
		Subtotal:           o.Subtotal,
		DiscountTotal:      o.DiscountTotal,
		TotalAmount:        o.TotalAmount,
		Status:             string(o.Status),
		ExpiresAt:          o.ExpiresAt,
		PaidAt:             o.PaidAt,
		CancelledAt:        o.CancelledAt,
		CreatedAt:          o.CreatedAt,
		UpdatedAt:          o.UpdatedAt,
		Items:              items,
		BatchAttendeeCount: o.BatchAttendeeCount,
		BatchContext:       toBatchContextDTO(o.BatchCtx),
	}
}

// toBatchContextDTO converts a domain BatchContext to its wire shape.
// Returns nil for non-batch orders.
func toBatchContextDTO(c *paymentdomain.BatchContext) *BatchContextDTO {
	if c == nil {
		return nil
	}
	return &BatchContextDTO{
		EventID:           c.EventID,
		AccountID:         c.AccountID,
		TemplateID:        c.TemplateID,
		PricePerCertMinor: c.PricePerCertMinor,
		EventTitle:        c.EventTitle,
		EventDate:         formatRFC3339(c.EventDate),
		EventDuration:     c.EventDuration,
		IssuerName:        c.IssuerName,
		AttendeeCount:     c.AttendeeCount,
	}
}

// formatRFC3339 returns the time in RFC3339 form, or "" for a zero
// time. Used so optional time fields serialize as omitted rather than
// "0001-01-01T00:00:00Z".
func formatRFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}