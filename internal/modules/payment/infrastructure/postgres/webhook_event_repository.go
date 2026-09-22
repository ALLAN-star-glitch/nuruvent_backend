// internal/modules/payment/infrastructure/postgres/webhook_event_repository.go

package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// WebhookEventRepository implements paymentdomain.WebhookEventRepository
// against Postgres.
type WebhookEventRepository struct {
	db *gorm.DB
}

// NewWebhookEventRepository constructs a WebhookEventRepository.
func NewWebhookEventRepository(db *gorm.DB) *WebhookEventRepository {
	return &WebhookEventRepository{db: db}
}

// ============================================================
// WRITE
// ============================================================

// RecordIfNew inserts a webhook event if the (provider, provider_event_id)
// pair is new. Returns ErrDuplicateWebhook when the event has already
// been recorded.
//
// The uniqueness is enforced by a database constraint, not by an
// application-level check. This means two concurrent deliveries of the
// same event resolve deterministically: one inserts, the other fails
// with a unique violation that we translate to ErrDuplicateWebhook.
func (r *WebhookEventRepository) RecordIfNew(ctx context.Context, e *paymentdomain.WebhookEvent) error {
	model := toWebhookEventModel(e)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: %s", paymentdomain.ErrDuplicateWebhook, err)
		}
		return translateError(err, "record webhook event")
	}
	return nil
}

// Update persists changes to a webhook event's processing status.
func (r *WebhookEventRepository) Update(ctx context.Context, e *paymentdomain.WebhookEvent) error {
	res := r.db.WithContext(ctx).
		Model(&WebhookEventModel{}).
		Where("id = ?", e.ID).
		Updates(map[string]any{
			"processed_at":     e.ProcessedAt,
			"processing_error": nullableString(e.ProcessingError),
		})

	if res.Error != nil {
		return translateError(res.Error, "update webhook event")
	}
	if res.RowsAffected == 0 {
		return paymentdomain.ErrWebhookEventNotFound
	}
	return nil
}

// ============================================================
// READ
// ============================================================

// FindByID returns a webhook event by its internal ID.
func (r *WebhookEventRepository) FindByID(ctx context.Context, id string) (*paymentdomain.WebhookEvent, error) {
	var model WebhookEventModel
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, paymentdomain.ErrWebhookEventNotFound
	}
	if err != nil {
		return nil, translateError(err, "find webhook event")
	}
	return toWebhookEventDomain(&model), nil
}

// FindUnprocessed returns webhook events that need processing. This
// includes events that failed (retryable) and events that were never
// attempted. Ordered oldest-first so backlogs drain in FIFO order.
//
// A caller (webhook worker) typically loops over batches:
//
//	for {
//	    events, _ := repo.FindUnprocessed(ctx, 100)
//	    if len(events) == 0 { break }
//	    for _, e := range events { process(e) }
//	}
func (r *WebhookEventRepository) FindUnprocessed(
	ctx context.Context,
	limit int,
) ([]*paymentdomain.WebhookEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	var models []WebhookEventModel
	err := r.db.WithContext(ctx).
		Where("processed_at IS NULL OR processing_error IS NOT NULL").
		Order("received_at ASC").
		Limit(limit).
		Find(&models).Error

	if err != nil {
		return nil, translateError(err, "find unprocessed webhook events")
	}

	events := make([]*paymentdomain.WebhookEvent, 0, len(models))
	for i := range models {
		events = append(events, toWebhookEventDomain(&models[i]))
	}
	return events, nil
}

// ============================================================
// COMPILE-TIME ASSERTION
// ============================================================

var _ paymentdomain.WebhookEventRepository = (*WebhookEventRepository)(nil)