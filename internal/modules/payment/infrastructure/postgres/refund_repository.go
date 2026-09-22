// internal/modules/payment/infrastructure/postgres/refund_repository.go

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// RefundRepository implements paymentdomain.RefundRepository against Postgres.
type RefundRepository struct {
	db *gorm.DB
}

// NewRefundRepository constructs a RefundRepository.
func NewRefundRepository(db *gorm.DB) *RefundRepository {
	return &RefundRepository{db: db}
}

// ============================================================
// WRITE
// ============================================================

// Create inserts a new refund.
func (r *RefundRepository) Create(ctx context.Context, refund *paymentdomain.Refund) error {
	model := toRefundModel(refund)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return translateError(err, "create refund")
	}
	return nil
}

// Update updates a refund's mutable fields. Amount, currency,
// payment_id, and actor_id are immutable.
func (r *RefundRepository) Update(ctx context.Context, refund *paymentdomain.Refund) error {
	res := r.db.WithContext(ctx).
		Model(&RefundModel{}).
		Where("id = ?", refund.ID).
		Updates(map[string]any{
			"status":             string(refund.Status),
			"provider_reference": nullableString(refund.ProviderReference),
			"failure_reason":     nullableString(refund.FailureReason),
			"completed_at":       refund.CompletedAt,
			"updated_at":         refund.UpdatedAt,
		})

	if res.Error != nil {
		return translateError(res.Error, "update refund")
	}
	if res.RowsAffected == 0 {
		return paymentdomain.ErrRefundNotFound
	}
	return nil
}

// ============================================================
// READ
// ============================================================

// FindByID returns a refund by its ID.
func (r *RefundRepository) FindByID(ctx context.Context, id string) (*paymentdomain.Refund, error) {
	var model RefundModel
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, paymentdomain.ErrRefundNotFound
	}
	if err != nil {
		return nil, translateError(err, "find refund")
	}
	return toRefundDomain(&model)
}

// ListByPayment returns all refunds for a payment, newest first.
//
// This is the primary lookup for computing the remaining refundable
// balance. Callers sum the succeeded and pending refunds from this
// list to enforce the "refunds cannot exceed payment amount" rule.
func (r *RefundRepository) ListByPayment(
	ctx context.Context,
	paymentID string,
) ([]*paymentdomain.Refund, error) {
	var models []RefundModel
	err := r.db.WithContext(ctx).
		Where("payment_id = ?", paymentID).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, translateError(err, "list refunds by payment")
	}

	refunds := make([]*paymentdomain.Refund, 0, len(models))
	for i := range models {
		refund, err := toRefundDomain(&models[i])
		if err != nil {
			return nil, err
		}
		refunds = append(refunds, refund)
	}
	return refunds, nil
}

// FindPendingBefore returns pending refunds created before the given
// time, ordered oldest-first. Used by a reconciliation job to time out
// refunds that never completed.
func (r *RefundRepository) FindPendingBefore(
	ctx context.Context,
	before time.Time,
	limit int,
) ([]*paymentdomain.Refund, error) {
	if limit <= 0 {
		limit = 100
	}

	var models []RefundModel
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", string(paymentdomain.PaymentStatusPending), before).
		Order("created_at ASC").
		Limit(limit).
		Find(&models).Error

	if err != nil {
		return nil, translateError(err, "find pending refunds")
	}

	refunds := make([]*paymentdomain.Refund, 0, len(models))
	for i := range models {
		refund, err := toRefundDomain(&models[i])
		if err != nil {
			return nil, err
		}
		refunds = append(refunds, refund)
	}
	return refunds, nil
}

// ============================================================
// COMPILE-TIME ASSERTION
// ============================================================

var _ paymentdomain.RefundRepository = (*RefundRepository)(nil)

// Silence unused import warning when fmt is only used in helpers below.
var _ = fmt.Sprintf