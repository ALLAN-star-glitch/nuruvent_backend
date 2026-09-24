// internal/modules/payment/service/create_order.go

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// DefaultOrderTTL is how long a pending order stays valid before it
// expires. Can be overridden at the call site if needed.
const DefaultOrderTTL = 30 * time.Minute

// ============================================================
// CREATE ORDER
// ============================================================

// CreateOrder builds a pending order from a registration.
//
// Steps:
//  1. Resolve the registration's pricing via the cross-module port.
//  2. Enforce ownership (authenticated actor or matching guest email).
//  3. Check for an existing pending order for the same registration.
//     If one exists, return it (idempotent).
//  4. Build the Order entity with items from the pricing snapshot.
//  5. Persist it.
//  6. Return it.
//
// The client supplies only the registration ID. All pricing, items,
// and identity come from the registration — the client cannot
// influence the order amount.
func (s *service) CreateOrder(
	ctx context.Context,
	cmd CreateOrderCommand,
) (*paymentdomain.Order, error) {
	if cmd.RegistrationID == "" {
		return nil, fmt.Errorf("registration_id is required")
	}

	now := s.deps.Clock.Now()

	// ------------------------------------------------------------
	// 1. Resolve pricing from the registration module.
	//    The snapshot carries UserID + GuestEmail, which we use for
	//    the ownership check below — no separate call needed.
	// ------------------------------------------------------------
	pricing, err := s.deps.RegistrationPricing.ResolvePricing(ctx, cmd.RegistrationID)
	if err != nil {
		return nil, fmt.Errorf("resolve registration pricing: %w", err)
	}
	if pricing == nil {
		return nil, fmt.Errorf("registration pricing not found for %s", cmd.RegistrationID)
	}

	// ------------------------------------------------------------
	// 2. Ownership check — BEFORE any side effects.
	//    Mirrors the rule in registration.GetByID.
	// ------------------------------------------------------------
	if err := assertRegistrationOwnership(cmd, pricing); err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 3. Duplicate check
	// ------------------------------------------------------------
	existing, err := s.deps.Orders.FindActiveByRegistration(ctx, cmd.RegistrationID)
	if err == nil && existing != nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, paymentdomain.ErrOrderNotFound) {
		return nil, fmt.Errorf("duplicate check: %w", err)
	}

	// ------------------------------------------------------------
	// 4. Build the Order entity
	// ------------------------------------------------------------
	items := make([]paymentdomain.OrderItem, 0, len(pricing.Items))
	for _, it := range pricing.Items {
		items = append(items, paymentdomain.OrderItem{
			TicketTypeID: it.TicketTypeID,
			Quantity:     it.Quantity,
			UnitPrice:    it.UnitPrice,
			Discount:     it.Discount,
			LineTotal:    it.UnitPrice*int64(it.Quantity) - it.Discount,
		})
	}

	orderID := s.deps.IDGenerator.NewID()

	order, err := paymentdomain.NewOrder(
		orderID,
		pricing.RegistrationID,
		pricing.UserID,
		pricing.GuestEmail,
		pricing.Currency,
		items,
		DefaultOrderTTL,
		now,
	)
	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 5. Persist
	// ------------------------------------------------------------
	if err := s.deps.Orders.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("persist order: %w", err)
	}

	return order, nil
}

// assertRegistrationOwnership enforces the same identity rule used by
// the registration module's GetByID:
//
//	authenticated actor → must own the registration
//	guest email         → must match the stored guest_email
//	neither             → unauthorized
//
// Note: UserID and GuestEmail are plain strings on RegistrationPricing;
// an empty value means "not set" for that identity path.
func assertRegistrationOwnership(
	cmd CreateOrderCommand,
	pricing *paymentdomain.RegistrationPricing,
) error {
	switch {
	case cmd.ActorID != "":
		if pricing.UserID == "" || pricing.UserID != cmd.ActorID {
			return paymentdomain.ErrForbidden
		}
		return nil

	case cmd.GuestEmail != "":
		if pricing.GuestEmail == "" ||
			!strings.EqualFold(pricing.GuestEmail, cmd.GuestEmail) {
			return paymentdomain.ErrForbidden
		}
		return nil

	default:
		return paymentdomain.ErrUnauthorized
	}
}

// ============================================================
// GET ORDER
// ============================================================

// GetOrder returns an order by ID. Read-only.
func (s *service) GetOrder(ctx context.Context, orderID string) (*paymentdomain.Order, error) {
	if orderID == "" {
		return nil, paymentdomain.ErrOrderNotFound
	}
	return s.deps.Orders.FindByID(ctx, orderID)
}