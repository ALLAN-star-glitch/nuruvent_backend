// internal/modules/payment/infrastructure/postgres/payment_repository.go

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// PaymentRepository implements paymentdomain.PaymentRepository against Postgres.
type PaymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository constructs a PaymentRepository.
func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// ============================================================
// WRITE
// ============================================================

// Create inserts a new payment. Enforces idempotency at the DB level:
// if the (order_id, idempotency_key) pair already exists, Postgres
// returns a unique violation which is translated to ErrDuplicatePayment.
func (r *PaymentRepository) Create(ctx context.Context, p *paymentdomain.Payment) error {
	model := toPaymentModel(p)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: %s", paymentdomain.ErrDuplicatePayment, err)
		}
		return translateError(err, "create payment")
	}
	return nil
}

// Update updates a payment's mutable fields. Amount, currency,
// idempotency_key, and order_id are immutable — the allowlist prevents
// accidental overwrites.
func (r *PaymentRepository) Update(ctx context.Context, p *paymentdomain.Payment) error {
	res := r.db.WithContext(ctx).
		Model(&PaymentModel{}).
		Where("id = ?", p.ID).
		Updates(map[string]any{
			"status":             string(p.Status),
			"provider_reference": nullableString(p.ProviderReference),
			"failure_reason":     nullableString(p.FailureReason),
			"redirect_url":       p.RedirectURL,
			"completed_at":       p.CompletedAt,
			"failed_at":          p.FailedAt,
			"updated_at":         p.UpdatedAt,
		})

	if res.Error != nil {
		return translateError(res.Error, "update payment")
	}
	if res.RowsAffected == 0 {
		return paymentdomain.ErrPaymentNotFound
	}
	return nil
}

// ============================================================
// READ
// ============================================================

// FindByID returns a payment by its internal ID.
func (r *PaymentRepository) FindByID(ctx context.Context, id string) (*paymentdomain.Payment, error) {
	var model PaymentModel
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, paymentdomain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, translateError(err, "find payment")
	}
	return toPaymentDomain(&model)
}

// FindByProviderReference looks up a payment by the provider's
// transaction ID. This is the primary entry point for webhooks and
// reconciliation, which only know the provider's reference — not our
// internal IDs.
func (r *PaymentRepository) FindByProviderReference(
	ctx context.Context,
	provider, reference string,
) (*paymentdomain.Payment, error) {
	if provider == "" || reference == "" {
		return nil, paymentdomain.ErrPaymentNotFound
	}

	var model PaymentModel
	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_reference = ?", provider, reference).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, paymentdomain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, translateError(err, "find payment by provider reference")
	}
	return toPaymentDomain(&model)
}

// FindByIdempotencyKey looks up a payment by (order_id, idempotency_key).
// Used to detect duplicate initiation requests — the caller returns the
// existing payment rather than creating a new one.
func (r *PaymentRepository) FindByIdempotencyKey(
	ctx context.Context,
	orderID, key string,
) (*paymentdomain.Payment, error) {
	if orderID == "" || key == "" {
		return nil, paymentdomain.ErrPaymentNotFound
	}

	var model PaymentModel
	err := r.db.WithContext(ctx).
		Where("order_id = ? AND idempotency_key = ?", orderID, key).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, paymentdomain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, translateError(err, "find payment by idempotency key")
	}
	return toPaymentDomain(&model)
}

// FindPendingExpired returns payments that are still pending but past
// their deadline, ordered oldest-first. Capped by limit.
func (r *PaymentRepository) FindPendingExpired(
	ctx context.Context,
	before time.Time,
	limit int,
) ([]*paymentdomain.Payment, error) {
	if limit <= 0 {
		limit = 100
	}

	var models []PaymentModel
	err := r.db.WithContext(ctx).
		Where("status = ? AND expires_at < ?", string(paymentdomain.PaymentStatusPending), before).
		Order("expires_at ASC").
		Limit(limit).
		Find(&models).Error

	if err != nil {
		return nil, translateError(err, "find pending expired payments")
	}

	payments := make([]*paymentdomain.Payment, 0, len(models))
	for i := range models {
		p, err := toPaymentDomain(&models[i])
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}

// FindByOrderID returns all payments for an order, newest first.
func (r *PaymentRepository) FindByOrderID(
	ctx context.Context,
	orderID string,
) ([]*paymentdomain.Payment, error) {
	var models []PaymentModel
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, translateError(err, "find payments by order")
	}

	payments := make([]*paymentdomain.Payment, 0, len(models))
	for i := range models {
		p, err := toPaymentDomain(&models[i])
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}


// FindPendingByOrder returns the current pending payment for an order,
// if one exists. Only one pending payment per order is allowed.
func (r *PaymentRepository) FindPendingByOrder(
	ctx context.Context,
	orderID string,
) (*paymentdomain.Payment, error) {
	var model PaymentModel
	err := r.db.WithContext(ctx).
		Where("order_id = ? AND status = ?",
			orderID,
			string(paymentdomain.PaymentStatusPending),
		).
		Order("created_at DESC").
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, paymentdomain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find pending payment: %w", err)
	}
	return toPaymentDomain(&model)
}

// ============================================================
// COMPILE-TIME ASSERTION
// ============================================================

var _ paymentdomain.PaymentRepository = (*PaymentRepository)(nil)