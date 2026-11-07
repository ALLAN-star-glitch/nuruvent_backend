// internal/modules/payment/infrastructure/postgres/order_repository.go

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// OrderRepository implements paymentdomain.OrderRepository against Postgres.
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository constructs an OrderRepository.
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// ============================================================
// WRITE
// ============================================================

// Create inserts a new order and its items in a single transaction.
func (r *OrderRepository) Create(ctx context.Context, o *paymentdomain.Order) error {
	model := toOrderModel(o)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return translateError(err, "create order")
	}
	return nil
}

// Update updates an order's scalar fields. Items are immutable once
// created — if items need to change, cancel the order and create a
// new one.
//
// Billing fields that change during the order's life:
//   - processing_fee_rate: snapshotted at initiate time once the
//     payment method is known (starts at 0, becomes 0.015 for M-Pesa
//     or 0.038 for card).
//
// Billing fields that are genuinely immutable after creation:
//   - platform_fee_rate, billed_account_id (set in NewOrder)
//   - subtotal, discount_total, total_amount, currency (pricing
//     snapshot from the registration)
//   - registration_id, user_id, guest_email (identity)
//   - expires_at, created_at (lifecycle)
//
// Settlement fields (settled_at, payout_ref) are written by the admin
// payout flow — add them here if that flow goes through Update rather
// than a dedicated Settle method.
func (r *OrderRepository) Update(ctx context.Context, o *paymentdomain.Order) error {
	res := r.db.WithContext(ctx).
		Model(&OrderModel{}).
		Where("id = ?", o.ID).
		Updates(map[string]any{
			"status":              string(o.Status),
			"paid_at":             o.PaidAt,
			"cancelled_at":        o.CancelledAt,
			"processing_fee_rate": o.ProcessingFeeRate, // ← new: set at initiate
			"updated_at":          o.UpdatedAt,
		})

	if res.Error != nil {
		return translateError(res.Error, "update order")
	}
	if res.RowsAffected == 0 {
		return paymentdomain.ErrOrderNotFound
	}
	return nil
}

// ============================================================
// READ
// ============================================================

// FindByID returns an order by its ID, with items preloaded and the
// latest payment status attached.
func (r *OrderRepository) FindByID(ctx context.Context, id string) (*paymentdomain.Order, error) {
	var model OrderModel
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("id = ?", id).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, paymentdomain.ErrOrderNotFound
	}
	if err != nil {
		return nil, translateError(err, "find order")
	}

	order, err := toOrderDomain(&model)
	if err != nil {
		return nil, err
	}

	// Fetch the most recent payment for this order and attach its
	// status to the order. This allows guests to poll GET /orders/:id
	// (which allows optionalAuth) to discover the payment outcome
	// without needing access to GET /payments/:id.
	var latestPayment PaymentModel
	err = r.db.WithContext(ctx).
		Where("order_id = ?", id).
		Order("created_at DESC").
		First(&latestPayment).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, translateError(err, "find latest payment for order")
	}
	if err == nil {
		order.PaymentStatus = paymentdomain.PaymentStatus(latestPayment.Status)
	}

	return order, nil
}

// FindActiveByRegistration returns the pending order for a registration,
// if any. Enforces the one-active-order-per-registration rule at the
// query level.
func (r *OrderRepository) FindActiveByRegistration(
	ctx context.Context,
	registrationID string,
) (*paymentdomain.Order, error) {
	var model OrderModel
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("registration_id = ? AND status = ?", registrationID, string(paymentdomain.OrderStatusPending)).
		Where("expires_at > NOW()").
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, paymentdomain.ErrOrderNotFound
	}
	if err != nil {
		return nil, translateError(err, "find active order")
	}
	return toOrderDomain(&model)
}

// FindExpired returns orders that are still pending but past their
// deadline, ordered oldest-first. Capped by limit.
func (r *OrderRepository) FindExpired(
	ctx context.Context,
	before time.Time,
	limit int,
) ([]*paymentdomain.Order, error) {
	if limit <= 0 {
		limit = 100
	}

	var models []OrderModel
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("status = ? AND expires_at < ?", string(paymentdomain.OrderStatusPending), before).
		Order("expires_at ASC").
		Limit(limit).
		Find(&models).Error

	if err != nil {
		return nil, translateError(err, "find expired orders")
	}

	orders := make([]*paymentdomain.Order, 0, len(models))
	for i := range models {
		o, err := toOrderDomain(&models[i])
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// FindActiveByCertificateContext returns the pending certificate order
// for a completion, if any.
func (r *OrderRepository) FindActiveByCertificateContext(
	ctx context.Context,
	contextType, contextID, recipientEmail string,
) (*paymentdomain.Order, error) {
	var model OrderModel
	err := r.db.WithContext(ctx).
		Where("kind = ? AND status = ?", "certificate", string(paymentdomain.OrderStatusPending)).
		Where("expires_at > NOW()").                              // ← ADD THIS LINE
		Where("certificate_context->>'context_type' = ?", contextType).
		Where("certificate_context->>'context_id' = ?", contextID).
		Where("certificate_context->>'recipient_email' = ?", recipientEmail).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, paymentdomain.ErrOrderNotFound
	}
	if err != nil {
		return nil, translateError(err, "find active certificate order")
	}
	return toOrderDomain(&model)
}

// ============================================================
// ERROR TRANSLATION
// ============================================================

// translateError converts a GORM/Postgres error into a domain error
// where appropriate. Preserves the underlying error with %w.
func translateError(err error, op string) error {
	if err == nil {
		return nil
	}

	// Postgres unique violation
	if isUniqueViolation(err) {
		return fmt.Errorf("%w: %s", paymentdomain.ErrDuplicateOrder, op)
	}

	// Postgres foreign key violation
	if isForeignKeyViolation(err) {
		return fmt.Errorf("%w: %s (foreign key)", paymentdomain.ErrOrderNotFound, op)
	}

	return fmt.Errorf("%s: %w", op, err)
}

// FindActiveByBatchEvent returns the pending certificate batch order
// for an event, if any.
func (r *OrderRepository) FindActiveByBatchEvent(
	ctx context.Context,
	eventID string,
) (*paymentdomain.Order, error) {
	var model OrderModel
	err := r.db.WithContext(ctx).
		Where("kind = ? AND status = ?", "certificate_batch", string(paymentdomain.OrderStatusPending)).
		Where("expires_at > NOW()").
		Where("batch_context->>'event_id' = ?", eventID).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, paymentdomain.ErrOrderNotFound
	}
	if err != nil {
		return nil, translateError(err, "find active batch order")
	}
	return toOrderDomain(&model)
}