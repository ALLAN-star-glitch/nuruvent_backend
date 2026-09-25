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
//  3. Inside a transaction:
//     a. Check for an existing pending order for the same
//        registration. If one exists, return it (idempotent).
//     b. Build the Order entity from the pricing snapshot.
//     c. Persist it.
//  4. Return it.
//
// The client supplies only the registration ID. All pricing, items,
// and identity come from the registration — the client cannot
// influence the order amount.
//
// Concurrency: the duplicate check and the insert share a single
// transaction, so two concurrent requests for the same registration
// cannot both create an order. The database's unique partial index on
// (registration_id) WHERE status = 'pending' is the final safety net;
// if it fires, the repository translates the Postgres 23505 into
// paymentdomain.ErrDuplicateOrder, and we re-fetch and return the
// winning order.
func (s *service) CreateOrder(
	ctx context.Context,
	cmd CreateOrderCommand,
) (*paymentdomain.Order, error) {
	if cmd.RegistrationID == "" {
		return nil, fmt.Errorf("registration_id is required")
	}

	// ------------------------------------------------------------
	// 1. Resolve pricing from the registration module.
	//    The snapshot carries UserID + GuestEmail, which we use for
	//    the ownership check below — no separate call needed.
	//
	//    This is a cross-module read. It happens outside the
	//    transaction because it may hit another module's tables, and
	//    we don't want to hold a payment-module transaction open
	//    across it.
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

	now := s.deps.Clock.Now()

	// ------------------------------------------------------------
	// 3. Everything that touches the DB goes inside Do.
	// ------------------------------------------------------------
	var order *paymentdomain.Order

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos paymentdomain.Repositories) error {
		// --------------------------------------------------------
		// 3a. Duplicate check inside the transaction.
		//     If a pending order already exists, return it. The
		//     transaction commits with no writes.
		// --------------------------------------------------------
		existing, err := repos.Orders.FindActiveByRegistration(ctx, cmd.RegistrationID)
		if err == nil && existing != nil {
			order = existing
			return nil
		}
		if err != nil && !errors.Is(err, paymentdomain.ErrOrderNotFound) {
			return fmt.Errorf("duplicate check: %w", err)
		}

		// --------------------------------------------------------
		// 3b. Build the Order entity.
		// --------------------------------------------------------
		order_items := make([]paymentdomain.OrderItem, 0, len(pricing.Items))
		for _, it := range pricing.Items {
			order_items = append(order_items, paymentdomain.OrderItem{
				TicketTypeID: it.TicketTypeID,
				Quantity:     it.Quantity,
				UnitPrice:    it.UnitPrice,
				Discount:     it.Discount,
				LineTotal:    it.UnitPrice*int64(it.Quantity) - it.Discount,
			})
		}

		newOrder, err := paymentdomain.NewOrder(
			s.deps.IDGenerator.NewID(),
			pricing.RegistrationID,
			pricing.UserID,
			pricing.GuestEmail,
			pricing.Currency,
			order_items,
			DefaultOrderTTL,
			now,
		)
		if err != nil {
			return err
		}

		// --------------------------------------------------------
		// 3c. Persist inside the transaction.
		// --------------------------------------------------------
		if err := repos.Orders.Create(ctx, newOrder); err != nil {
			return fmt.Errorf("persist order: %w", err)
		}

		// Future: enqueue an outbox event here for "order.created".
		// Example:
		//   if err := repos.Outbox.Enqueue(ctx, outbox.Event{
		//       Type:    "order.created",
		//       Payload: marshal(newOrder),
		//   }); err != nil {
		//       return err
		//   }

		order = newOrder
		return nil
	})

	if txErr != nil {
		// ----------------------------------------------------------
		// Unique-constraint safety net.
		//
		// Two concurrent requests can both pass the duplicate check
		// inside their own transactions and both attempt the INSERT.
		// The partial unique index on (registration_id)
		// WHERE status = 'pending' rejects the second one with
		// Postgres SQLSTATE 23505, which the repository translates
		// into paymentdomain.ErrDuplicateOrder.
		//
		// When that happens, re-fetch and return the winner so the
		// caller sees idempotent behavior instead of a 500.
		// ----------------------------------------------------------
		if errors.Is(txErr, paymentdomain.ErrDuplicateOrder) {
			existing, findErr := s.deps.Orders.FindActiveByRegistration(ctx, cmd.RegistrationID)
			if findErr == nil && existing != nil {
				return existing, nil
			}
			// Fall through: if we still can't find it, return the
			// original error so the caller sees something real.
		}
		return nil, txErr
	}

	// ------------------------------------------------------------
	// 4. Non-transactional side effects go AFTER commit.
	//
	//    Notifications, emails, webhooks to other systems — none of
	//    these belong inside the transaction. If they fail, the
	//    order is still valid; the notification can be retried.
	//
	//    For now there are none, but this is where they'd go:
	//      _ = s.deps.Notifier.OrderCreated(ctx, order)
	// ------------------------------------------------------------

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