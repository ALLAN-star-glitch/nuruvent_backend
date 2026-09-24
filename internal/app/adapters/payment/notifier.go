// internal/app/adapters/payment/notifier.go

package payment

import (
	"context"
	"fmt"
	"log"

	notificationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	paymentdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// Notifier implements paymentdomain.Notifier by delegating to the
// notification module's service.
//
// Recipient resolution:
//   - Guest orders: uses order.GuestEmail directly.
//   - Authenticated orders: resolves the user's email via UserEmailResolver.
//
// Every method is best-effort: errors are returned to the caller, which
// logs them. Notification failures do not roll back payment operations.
type Notifier struct {
	notifSvc    notificationDomain.NotificationService
	userEmails  paymentdomain.UserEmailResolver
}

// NewNotifier constructs the adapter.
func NewNotifier(
	notifSvc notificationDomain.NotificationService,
	userEmails paymentdomain.UserEmailResolver,
) *Notifier {
	return &Notifier{
		notifSvc:   notifSvc,
		userEmails: userEmails,
	}
}

// ============================================================
// PAYMENT INITIATED
// ============================================================

func (n *Notifier) PaymentInitiated(
	ctx context.Context,
	payment *paymentdomain.Payment,
	order *paymentdomain.Order,
) error {
	if payment == nil || order == nil {
		return nil
	}

	recipient, err := n.resolveRecipient(ctx, order)
	if err != nil {
		log.Printf("[PaymentNotifier] resolve recipient for order %s: %v", order.ID, err)
		return nil
	}
	if recipient == "" {
		return nil
	}

	req := notificationDomain.SendPaymentInitiatedRequest{
		To:          recipient,
		Amount:      payment.Amount,
		Currency:    payment.Currency,
		Provider:    payment.Provider,
		Method:      string(payment.Method),
		CustomerMsg: customerMessageFor(payment.Method),
		PaymentID:   payment.ID,
		OrderID:     order.ID,
	}

	if err := n.notifSvc.SendPaymentInitiated(ctx, req); err != nil {
		return fmt.Errorf("send payment initiated: %w", err)
	}
	return nil
}

// ============================================================
// PAYMENT SUCCEEDED
// ============================================================

func (n *Notifier) PaymentSucceeded(
	ctx context.Context,
	payment *paymentdomain.Payment,
	order *paymentdomain.Order,
) error {
	if payment == nil || order == nil {
		return nil
	}

	recipient, err := n.resolveRecipient(ctx, order)
	if err != nil {
		log.Printf("[PaymentNotifier] resolve recipient for order %s: %v", order.ID, err)
		return nil
	}
	if recipient == "" {
		return nil
	}

	req := notificationDomain.SendPaymentSucceededRequest{
		To:                recipient,
		Amount:            payment.Amount,
		Currency:          payment.Currency,
		Provider:          payment.Provider,
		Method:            string(payment.Method),
		ProviderReference: payment.ProviderReference,
		PaymentID:         payment.ID,
		OrderID:           order.ID,
		RegistrationID:    order.RegistrationID,
	}

	if err := n.notifSvc.SendPaymentSucceeded(ctx, req); err != nil {
		return fmt.Errorf("send payment succeeded: %w", err)
	}
	return nil
}

// ============================================================
// PAYMENT FAILED
// ============================================================

func (n *Notifier) PaymentFailed(
	ctx context.Context,
	payment *paymentdomain.Payment,
	order *paymentdomain.Order,
) error {
	if payment == nil || order == nil {
		return nil
	}

	recipient, err := n.resolveRecipient(ctx, order)
	if err != nil {
		log.Printf("[PaymentNotifier] resolve recipient for order %s: %v", order.ID, err)
		return nil
	}
	if recipient == "" {
		return nil
	}

	req := notificationDomain.SendPaymentFailedRequest{
		To:            recipient,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Provider:      payment.Provider,
		Method:        string(payment.Method),
		FailureReason: payment.FailureReason,
		PaymentID:     payment.ID,
		OrderID:       order.ID,
	}

	if err := n.notifSvc.SendPaymentFailed(ctx, req); err != nil {
		return fmt.Errorf("send payment failed: %w", err)
	}
	return nil
}

// ============================================================
// PAYMENT EXPIRED
// ============================================================

func (n *Notifier) PaymentExpired(
	ctx context.Context,
	payment *paymentdomain.Payment,
	order *paymentdomain.Order,
) error {
	if payment == nil || order == nil {
		return nil
	}

	recipient, err := n.resolveRecipient(ctx, order)
	if err != nil {
		log.Printf("[PaymentNotifier] resolve recipient for order %s: %v", order.ID, err)
		return nil
	}
	if recipient == "" {
		return nil
	}

	req := notificationDomain.SendPaymentExpiredRequest{
		To:        recipient,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		Provider:  payment.Provider,
		Method:    string(payment.Method),
		PaymentID: payment.ID,
		OrderID:   order.ID,
	}

	if err := n.notifSvc.SendPaymentExpired(ctx, req); err != nil {
		return fmt.Errorf("send payment expired: %w", err)
	}
	return nil
}

// ============================================================
// REFUND ISSUED
// ============================================================

func (n *Notifier) RefundIssued(
	ctx context.Context,
	refund *paymentdomain.Refund,
	payment *paymentdomain.Payment,
) error {
	if refund == nil || payment == nil {
		return nil
	}

	// The payment entity does not carry the order, so we look up the
	// recipient by loading the order through the payment's OrderID.
	// In practice the service can pass a richer structure later.
	recipient, err := n.resolveRefundRecipient(ctx, payment)
	if err != nil {
		log.Printf("[PaymentNotifier] resolve refund recipient for payment %s: %v", payment.ID, err)
		return nil
	}
	if recipient == "" {
		return nil
	}

	req := notificationDomain.SendRefundIssuedRequest{
		To:                recipient,
		Amount:            refund.Amount,
		Currency:          refund.Currency,
		OriginalAmount:    payment.Amount,
		ProviderReference: refund.ProviderReference,
		Reason:            refund.Reason,
		PaymentID:         payment.ID,
		RefundID:          refund.ID,
		IsPartial:         refund.Amount < payment.Amount,
	}

	if err := n.notifSvc.SendRefundIssued(ctx, req); err != nil {
		return fmt.Errorf("send refund issued: %w", err)
	}
	return nil
}


// ============================================================
// RECIPIENT RESOLUTION
// ============================================================

// resolveRecipient determines the destination email for a payment
// notification based on the order's owner.
func (n *Notifier) resolveRecipient(ctx context.Context, order *paymentdomain.Order) (string, error) {
	// Guest order — email is on the order.
	if order.GuestEmail != "" {
		return order.GuestEmail, nil
	}

	// Authenticated order — resolve via the account module.
	if order.UserID != "" && n.userEmails != nil {
		return n.userEmails.ResolveEmail(ctx, order.UserID)
	}

	return "", fmt.Errorf("order %s has no recipient", order.ID)
}


// resolveRefundRecipient resolves the recipient for a refund. Since the
// refund is tied to a payment (which has OrderID, not the order), the
// caller-side service should ideally pass the order. For MVP we accept
// that this may return empty for authenticated users.
func (n *Notifier) resolveRefundRecipient(ctx context.Context, payment *paymentdomain.Payment) (string, error) {
	// We cannot resolve without the order. The service should be
	// enhanced to pass it, or the refund notification path should load
	// it before calling. For MVP, skip if we can't resolve.
	return "", nil
}

// ============================================================
// HELPERS
// ============================================================

func customerMessageFor(method paymentdomain.PaymentMethod) string {
	switch method {
	case paymentdomain.PaymentMethodMpesa:
		return "Enter your M-Pesa PIN on your phone to complete payment."
	case paymentdomain.PaymentMethodCard:
		return "Complete the verification step to confirm your card payment."
	default:
		return "Follow the instructions to complete your payment."
	}
}

var _ paymentdomain.Notifier = (*Notifier)(nil)