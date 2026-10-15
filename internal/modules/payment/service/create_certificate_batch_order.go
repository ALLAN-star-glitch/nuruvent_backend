// internal/modules/payment/service/create_certificate_batch_order.go

package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// CreateCertificateBatchOrder creates a pending order for a
// certificate batch. The organizer pays; attendees receive their
// certificates after payment succeeds.
//
// Idempotent per event: a repeat call returns the existing pending
// order for the same event.
func (s *service) CreateCertificateBatchOrder(
	ctx context.Context,
	cmd CreateCertificateBatchOrderCommand,
) (*paymentdomain.Order, error) {
	// ------------------------------------------------------------
	// 1. Validate input.
	// ------------------------------------------------------------
	if cmd.EventID == "" {
		return nil, fmt.Errorf("event_id is required")
	}
	if cmd.AccountID == "" {
		return nil, fmt.Errorf("account_id is required")
	}
	if cmd.TemplateID == "" {
		return nil, fmt.Errorf("template_id is required")
	}
	if cmd.PricePerCertMinor <= 0 {
		return nil, fmt.Errorf("price_per_cert_minor must be positive")
	}
	if len(cmd.Recipients) == 0 {
		return nil, fmt.Errorf("at least one recipient is required")
	}
	if cmd.Currency == "" {
		cmd.Currency = "KES"
	}

	// ------------------------------------------------------------
	// 2. Compute total.
	// ------------------------------------------------------------
	total := cmd.PricePerCertMinor * int64(len(cmd.Recipients))

	// ------------------------------------------------------------
	// 3. Build batch context (frozen snapshot).
	// ------------------------------------------------------------
	recipients := make([]paymentdomain.BatchRecipient, 0, len(cmd.Recipients))
	for _, r := range cmd.Recipients {
		recipients = append(recipients, paymentdomain.BatchRecipient{
			UserID:      r.UserID,
			Name:        r.Name,
			Email:       r.Email,
			CompletedAt: r.CompletedAt,
		})
	}

	batchCtx := paymentdomain.BatchContext{
		EventID:           cmd.EventID,
		AccountID:         cmd.AccountID,
		TemplateID:        cmd.TemplateID,
		PricePerCertMinor: cmd.PricePerCertMinor,
		EventTitle:        cmd.EventTitle,
		EventDate:         cmd.EventDate,
		EventDuration:     cmd.EventDuration,
		IssuerName:        cmd.IssuerName,
		AttendeeCount:     len(recipients),
		Recipients:        recipients,
	}

	// ------------------------------------------------------------
	// 4. Idempotency check + create inside a transaction.
	// ------------------------------------------------------------
	now := s.deps.Clock.Now()

	var order *paymentdomain.Order

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos paymentdomain.Repositories) error {
		// Duplicate check.
		existing, err := repos.Orders.FindActiveByBatchEvent(ctx, cmd.EventID)
		if err == nil && existing != nil {
			order = existing
			return nil
		}
		if err != nil && !errors.Is(err, paymentdomain.ErrOrderNotFound) {
			return fmt.Errorf("duplicate check: %w", err)
		}

		newOrder, err := paymentdomain.NewCertificateBatchOrder(
			s.deps.IDGenerator.NewID(),
			batchCtx,
			cmd.Currency,
			total,
			DefaultOrderTTL,
			now,
			cmd.AccountID,
		)
		if err != nil {
			return err
		}

		if err := repos.Orders.Create(ctx, newOrder); err != nil {
			return fmt.Errorf("persist order: %w", err)
		}

		order = newOrder
		return nil
	})

	if txErr != nil {
		// Race: another request created the batch order concurrently.
		// Re-fetch and return the winner.
		if errors.Is(txErr, paymentdomain.ErrDuplicateOrder) {
			existing, findErr := s.deps.Orders.FindActiveByBatchEvent(ctx, cmd.EventID)
			if findErr == nil && existing != nil {
				return existing, nil
			}
		}
		return nil, txErr
	}

	return order, nil
}