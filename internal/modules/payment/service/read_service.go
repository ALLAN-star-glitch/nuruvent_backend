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

	return s.deps.Payments.AggregateForAccount(ctx, f)
}