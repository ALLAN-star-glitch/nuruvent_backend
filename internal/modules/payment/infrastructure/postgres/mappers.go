// internal/modules/payment/infrastructure/postgres/mappers.go

package postgres

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// ============================================================
// ORDER
// ============================================================

// toOrderDomain converts an OrderModel to a domain Order.
func toOrderDomain(m *OrderModel) (*paymentdomain.Order, error) {
	items, err := toOrderItemsDomain(m.Items)
	if err != nil {
		return nil, err
	}

	status, err := paymentdomain.ParseOrderStatus(m.Status)
	if err != nil {
		return nil, err
	}

	return paymentdomain.HydrateOrder(
		m.ID,
		m.RegistrationID,
		derefString(m.UserID),
		derefString(m.GuestEmail),
		m.Currency,
		items,
		m.Subtotal,
		m.DiscountTotal,
		m.TotalAmount,
		status,
		m.ExpiresAt,
		m.CreatedAt,
		m.UpdatedAt,
		m.PaidAt,
		m.CancelledAt,
	), nil
}

// toOrderModel converts a domain Order to an OrderModel.
func toOrderModel(o *paymentdomain.Order) *OrderModel {
	model := &OrderModel{
		ID:             o.ID,
		RegistrationID: o.RegistrationID,
		UserID:         nullableString(o.UserID),
		GuestEmail:     nullableString(o.GuestEmail),
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
	}

	if len(o.Items) > 0 {
		model.Items = toOrderItemsModel(o.Items, o.ID)
	}

	return model
}

// ============================================================
// ORDER ITEM
// ============================================================

func toOrderItemDomain(m *OrderItemModel) paymentdomain.OrderItem {
	return paymentdomain.OrderItem{
		TicketTypeID: m.TicketTypeID,
		Quantity:     m.Quantity,
		UnitPrice:    m.UnitPrice,
		Discount:     m.Discount,
		LineTotal:    m.LineTotal,
	}
}

func toOrderItemsDomain(models []OrderItemModel) ([]paymentdomain.OrderItem, error) {
	items := make([]paymentdomain.OrderItem, 0, len(models))
	for i := range models {
		items = append(items, toOrderItemDomain(&models[i]))
	}
	return items, nil
}

func toOrderItemModel(item paymentdomain.OrderItem, orderID string) *OrderItemModel {
	return &OrderItemModel{
		OrderID:      orderID,
		TicketTypeID: item.TicketTypeID,
		Quantity:     item.Quantity,
		UnitPrice:    item.UnitPrice,
		Discount:     item.Discount,
		LineTotal:    item.LineTotal,
	}
}

func toOrderItemsModel(items []paymentdomain.OrderItem, orderID string) []OrderItemModel {
	models := make([]OrderItemModel, 0, len(items))
	for _, it := range items {
		models = append(models, *toOrderItemModel(it, orderID))
	}
	return models
}

// ============================================================
// PAYMENT
// ============================================================

func toPaymentDomain(m *PaymentModel) (*paymentdomain.Payment, error) {
	method, err := paymentdomain.ParsePaymentMethod(m.Method)
	if err != nil {
		return nil, err
	}

	status, err := paymentdomain.ParsePaymentStatus(m.Status)
	if err != nil {
		return nil, err
	}

	return paymentdomain.HydratePayment(
    m.ID,
    m.OrderID,
    m.Provider,
    method,
    m.Amount,
    m.Currency,
    status,
    m.IdempotencyKey,
    derefString(m.ProviderReference),
    derefString(m.FailureReason),     // 👈 10th: failureReason
    m.RedirectURL,                    // 👈 11th: redirectURL
    m.InitiatedAt,
    m.CompletedAt,
    m.FailedAt,
    m.ExpiresAt,
    m.CreatedAt,
    m.UpdatedAt,
), nil
}




func toPaymentModel(p *paymentdomain.Payment) *PaymentModel {
	return &PaymentModel{
		ID:                p.ID,
		OrderID:           p.OrderID,
		Provider:          p.Provider,
		Method:            string(p.Method),
		Amount:            p.Amount,
		Currency:          p.Currency,
		Status:            string(p.Status),
		IdempotencyKey:    p.IdempotencyKey,
		ProviderReference: nullableString(p.ProviderReference),
		RedirectURL: p.RedirectURL,
		FailureReason:     nullableString(p.FailureReason),
		InitiatedAt:       p.InitiatedAt,
		ExpiresAt:         p.ExpiresAt,
		CompletedAt:       p.CompletedAt,
		FailedAt:          p.FailedAt,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,

	}
}

// ============================================================
// REFUND
// ============================================================

func toRefundDomain(m *RefundModel) (*paymentdomain.Refund, error) {
	status, err := paymentdomain.ParsePaymentStatus(m.Status)
	if err != nil {
		return nil, err
	}

	return paymentdomain.HydrateRefund(
		m.ID,
		m.PaymentID,
		m.Amount,
		m.Currency,
		derefString(m.Reason),
		m.ActorID,
		derefString(m.ProviderReference),
		status,
		derefString(m.FailureReason),
		m.CreatedAt,
		m.UpdatedAt,
		m.CompletedAt,
	), nil
}

func toRefundModel(r *paymentdomain.Refund) *RefundModel {
	return &RefundModel{
		ID:                r.ID,
		PaymentID:         r.PaymentID,
		Amount:            r.Amount,
		Currency:          r.Currency,
		Reason:            nullableString(r.Reason),
		ActorID:           r.ActorID,
		ProviderReference: nullableString(r.ProviderReference),
		Status:            string(r.Status),
		FailureReason:     nullableString(r.FailureReason),
		CompletedAt:       r.CompletedAt,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

// ============================================================
// WEBHOOK EVENT
// ============================================================

func toWebhookEventDomain(m *WebhookEventModel) *paymentdomain.WebhookEvent {
	return paymentdomain.HydrateWebhookEvent(
		m.ID,
		m.Provider,
		m.ProviderEventID,
		m.SignatureValid,
		m.Payload,
		m.ReceivedAt,
		m.ProcessedAt,
		derefString(m.ProcessingError),
	)
}

func toWebhookEventModel(w *paymentdomain.WebhookEvent) *WebhookEventModel {
	return &WebhookEventModel{
		ID:              w.ID,
		Provider:        w.Provider,
		ProviderEventID: w.ProviderEventID,
		SignatureValid:  w.SignatureValid,
		Payload:         w.Payload,
		ReceivedAt:      w.ReceivedAt,
		ProcessedAt:     w.ProcessedAt,
		ProcessingError: nullableString(w.ProcessingError),
	}
}

// ============================================================
// HELPERS
// ============================================================

// nullableString converts an empty string to nil for DB storage.
// The DB treats NULL as "not set"; empty strings are ambiguous.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// derefString converts a nullable string to a plain string.
// Returns "" when the pointer is nil.
func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
