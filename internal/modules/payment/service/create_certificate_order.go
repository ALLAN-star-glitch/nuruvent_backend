// internal/modules/payment/service/create_certificate_order.go

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// CreateCertificateOrder builds a pending order for a certificate
// purchase.
//
// Idempotent per (context_type, context_id, recipient_email) — a
// repeat call returns the existing pending order.
//
// Ownership: the caller must supply either ActorID (the buyer's user
// ID) or GuestEmail (matching the recipient's email). Enforced here
// because there is no registration to cross-check against.
func (s *service) CreateCertificateOrder(
	ctx context.Context,
	cmd CreateCertificateOrderCommand,
) (*paymentdomain.Order, error) {
	// ------------------------------------------------------------
	// 1. Validate input.
	// ------------------------------------------------------------
	if cmd.ContextType == "" {
		return nil, fmt.Errorf("context_type is required")
	}
	if cmd.ContextID == "" {
		return nil, fmt.Errorf("context_id is required")
	}
	if cmd.RecipientName == "" {
		return nil, fmt.Errorf("recipient_name is required")
	}
	if cmd.RecipientEmail == "" {
		return nil, fmt.Errorf("recipient_email is required")
	}
	if cmd.AccountID == "" {
		return nil, fmt.Errorf("account_id is required")
	}
	if cmd.AmountMinor <= 0 {
		return nil, fmt.Errorf("amount_minor must be positive")
	}
	if cmd.Currency == "" {
		cmd.Currency = "KES"
	}

	// ------------------------------------------------------------
	// 2. Ownership check.
	// ------------------------------------------------------------
	if err := assertCertificateOrderOwnership(cmd); err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 3. Idempotency + create inside a transaction.
	// ------------------------------------------------------------
	now := s.deps.Clock.Now()

	certCtx := paymentdomain.CertificateContext{
		ContextType:     cmd.ContextType,
		ContextID:       cmd.ContextID,
		TemplateID:      cmd.TemplateID,
		RecipientName:   cmd.RecipientName,
		RecipientEmail:  strings.ToLower(strings.TrimSpace(cmd.RecipientEmail)),
		RecipientUserID: cmd.RecipientUserID,
		CompletedAt:     cmd.CompletedAt,
		AccountID:       cmd.AccountID,
	}

	var order *paymentdomain.Order

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos paymentdomain.Repositories) error {
		// Duplicate check.
		existing, err := repos.Orders.FindActiveByCertificateContext(
			ctx, cmd.ContextType, cmd.ContextID, certCtx.RecipientEmail,
		)
		if err == nil && existing != nil {
			order = existing
			return nil
		}
		if err != nil && !errors.Is(err, paymentdomain.ErrOrderNotFound) {
			return fmt.Errorf("duplicate check: %w", err)
		}

		// Build the order.
		newOrder, err := paymentdomain.NewCertificateOrder(
			s.deps.IDGenerator.NewID(),
			certCtx,
			cmd.ActorID,
			cmd.GuestEmail,
			cmd.Currency,
			cmd.AmountMinor,
			DefaultOrderTTL,
			now,
			cmd.AccountID,
			s.deps.Config.PaymentBilling.PlatformFeeRate,
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
		// Unique-constraint race: another request created the same
		// certificate order concurrently. Re-fetch and return the
		// winner for idempotent behavior.
		if errors.Is(txErr, paymentdomain.ErrDuplicateOrder) {
			existing, findErr := s.deps.Orders.FindActiveByCertificateContext(
				ctx, cmd.ContextType, cmd.ContextID, certCtx.RecipientEmail,
			)
			if findErr == nil && existing != nil {
				return existing, nil
			}
		}
		return nil, txErr
	}

	return order, nil
}

// assertCertificateOrderOwnership enforces the identity rule for
// certificate orders. Unlike registration orders, there is no
// registration to cross-check — the caller asserts identity directly.
//
// Exactly one of ActorID / GuestEmail must be set.
func assertCertificateOrderOwnership(cmd CreateCertificateOrderCommand) error {
	switch {
	case cmd.ActorID != "" && cmd.GuestEmail != "":
		return fmt.Errorf("%w: both actor and guest identity set",
			paymentdomain.ErrIdentityRequired)
	case cmd.ActorID != "":
		return nil
	case cmd.GuestEmail != "":
		// Guest email must match the recipient email.
		if !strings.EqualFold(cmd.GuestEmail, cmd.RecipientEmail) {
			return paymentdomain.ErrForbidden
		}
		return nil
	default:
		return paymentdomain.ErrUnauthorized
	}
}