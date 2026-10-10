// internal/modules/payment/service/read_service.go

package service

import (
	"context"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// ============================================================
// LIST PAYMENTS
// ============================================================

// ListPayments returns a paginated ledger of payments for one
// account. The account is the "billed account" — the snapshot taken
// at order creation, immutable thereafter.
//
// This is a read-only operation. It performs no writes, dispatches
// no notifications, and does not touch the unit of work.
//
// The fee breakdown (platform fee, processing fee, net to organizer)
// is computed in the repository's SQL, so the numbers here exactly
// match the totals from GetPaymentStats.
func (s *service) ListPayments(
	ctx context.Context,
	cmd ListPaymentsCommand,
) (*ListPaymentsResult, error) {
	if cmd.BilledAccountID == "" {
		return nil, fmt.Errorf("billed account id is required")
	}

	f := paymentdomain.ListPaymentsFilter{
		BilledAccountID: cmd.BilledAccountID,
		Page:            cmd.Page,
		PageSize:        cmd.PageSize,
		Search:          cmd.Search,
		Status:          cmd.Status,
		Method:          cmd.Method,
		EventID:         cmd.EventID,
		DateFrom:        cmd.DateFrom,
		DateTo:          cmd.DateTo,
		SortBy:          cmd.SortBy,
		SortOrder:       cmd.SortOrder,
	}

	if err := s.applyTeamFilter(ctx, &f, cmd.ActorID, cmd.TeamID, cmd.Scope); err != nil {
		return nil, err
	}

	rows, total, err := s.deps.Payments.ListForAccount(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}

	// Normalise the echoed pagination back to the caller so the
	// response matches what the repository actually did.
	page := cmd.Page
	if page < 1 {
		page = 1
	}
	pageSize := cmd.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return &ListPaymentsResult{
		Payments: rows,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ============================================================
// PAYMENT STATS
// ============================================================

// GetPaymentStats returns aggregate statistics for one account,
// honouring the same filters as ListPayments minus pagination.
//
// Read-only. The fee aggregates are computed in the repository's
// SQL, sharing the exact expressions used by ListPayments.
func (s *service) GetPaymentStats(
	ctx context.Context,
	cmd PaymentStatsCommand,
) (*paymentdomain.PaymentStats, error) {
	if cmd.BilledAccountID == "" {
		return nil, fmt.Errorf("billed account id is required")
	}

	f := paymentdomain.ListPaymentsFilter{
		BilledAccountID: cmd.BilledAccountID,
		EventID:         cmd.EventID,
		DateFrom:        cmd.DateFrom,
		DateTo:          cmd.DateTo,
	}

	if err := s.applyTeamFilter(ctx, &f, cmd.ActorID, cmd.TeamID, cmd.Scope); err != nil {
		return nil, err
	}

	return s.deps.Payments.AggregateForAccount(ctx, f)
}

// applyTeamFilter expands a team ID into event IDs and, when
// non-empty, narrows the filter to those events.
//
// Behaviour:
//   - Scope != "team" or TeamID == ""  → no-op, filter unchanged.
//   - TeamID set but ActorID empty     → ErrTeamAccessDenied (can't authorize).
//   - Team has zero events             → filter set to match nothing.
//   - Team not found or no access      → ErrTeamAccessDenied.
//
// The events service enforces permissions; this helper just
// propagates its errors.
func (s *service) applyTeamFilter(
	ctx context.Context,
	f *paymentdomain.ListPaymentsFilter,
	actorID, teamID, scope string,
) error {
	if scope != "team" || teamID == "" {
		return nil
	}
	if actorID == "" {
		return paymentdomain.ErrTeamAccessDenied
	}
	if s.deps.EventsReader == nil {
		return fmt.Errorf("events reader not configured")
	}

	eventIDs, err := s.deps.EventsReader.ListEventIDsByTeam(ctx, actorID, teamID)
	if err != nil {
		// The adapter translates the events domain's ErrForbidden
		// into paymentdomain.ErrTeamAccessDenied. Anything else is
		// propagated as-is.
		return err
	}

	if len(eventIDs) == 0 {
		// No events in the team — force an empty result. Using a
		// sentinel that can't match anything is simpler than
		// threading a separate "no results" flag through the repo.
		f.EventIDs = []string{"00000000-0000-0000-0000-000000000000"}
		return nil
	}

	f.EventIDs = eventIDs
	return nil
}