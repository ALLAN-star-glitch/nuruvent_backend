// internal/modules/payment/infrastructure/postgres/payment_repository.go

package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
// READ MODELS — LEDGER + STATS
// ============================================================

// ListForAccount returns a paginated payment ledger scoped to one
// account, along with the total count of matching rows (before
// pagination) so the caller can compute total pages.
//
// Account ownership is resolved through orders.billed_account_id, the
// immutable snapshot taken at order creation.
//
// The fee breakdown is computed in SQL:
//   platform_fee    = payment.amount × orders.platform_fee_rate
//   processing_fee  = payment.amount × payments.processing_fee_rate
//   net_to_organizer = amount − platform_fee − processing_fee
//
// Uses raw SQL deliberately — GORM's Table().Select().Scan() chain
// has silently dropped aliased columns on this codebase before.
func (r *PaymentRepository) ListForAccount(
	ctx context.Context,
	f paymentdomain.ListPaymentsFilter,
) ([]*paymentdomain.PaymentListRow, int, error) {
	if f.BilledAccountID == "" {
		return []*paymentdomain.PaymentListRow{}, 0, nil
	}

	// Normalise pagination.
	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Build the WHERE clause + args once; count and data share it.
	where := []string{
		"o.billed_account_id = ?",
		"o.deleted_at IS NULL",
		"r.deleted_at IS NULL",
		"e.deleted_at IS NULL",
	}
	args := []any{f.BilledAccountID}

	if f.Status != "" {
		where = append(where, "p.status = ?")
		args = append(args, f.Status)
	}
	if f.Method != "" {
		where = append(where, "p.method = ?")
		args = append(args, f.Method)
	}
	if f.EventID != "" {
		where = append(where, "e.id = ?")
		args = append(args, f.EventID)
	}
	if f.DateFrom != nil {
		where = append(where, "p.created_at >= ?")
		args = append(args, *f.DateFrom)
	}
	if f.DateTo != nil {
		where = append(where, "p.created_at <= ?")
		args = append(args, *f.DateTo)
	}
	if s := strings.TrimSpace(f.Search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		where = append(where,
			"(LOWER(COALESCE(u.name, r.guest_name, '')) LIKE ? "+
				"OR LOWER(COALESCE(u.email, r.guest_email, '')) LIKE ? "+
				"OR LOWER(COALESCE(p.provider_reference, '')) LIKE ? "+
				"OR LOWER(COALESCE(e.display_name, e.name, '')) LIKE ?)")
		args = append(args, like, like, like, like)
	}

	whereSQL := strings.Join(where, " AND ")

	// ------------------------------------------------------------
	// COUNT — total matching rows (before pagination)
	// ------------------------------------------------------------
	countSQL := `
		SELECT COUNT(*)
		FROM payments p
		JOIN orders o               ON o.id = p.order_id
		JOIN registrations r        ON r.id = o.registration_id
		JOIN event_registrations er ON er.registration_id = r.id
		JOIN events e               ON e.id = er.event_id
		LEFT JOIN users u           ON u.id = r.user_id
		WHERE ` + whereSQL

	var total int64
	if err := r.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count payments: %w", err)
	}

	// ------------------------------------------------------------
	// ORDER
	// ------------------------------------------------------------
	orderClause := buildPaymentOrderClause(f.SortBy, f.SortOrder)

	// ------------------------------------------------------------
	// DATA
	// ------------------------------------------------------------
	// SQL aliases are snake_case and match the domain struct fields
	// by GORM's convention — no tags on PaymentListRow.
	//
	// The fee math uses integer casts so Postgres stays in bigint
	// throughout: (numeric * numeric)::bigint truncates toward zero,
	// which is what we want.
	dataSQL := `
		SELECT
			p.id,
			p.order_id,
			p.provider,
			p.method,
			COALESCE(p.provider_reference, '')  AS provider_reference,
			p.amount,
			p.currency,
			p.status,

			-- Fee breakdown
			(p.amount * o.platform_fee_rate)::bigint                    AS platform_fee,
			(p.amount * p.processing_fee_rate)::bigint                  AS processing_fee,
			(p.amount
				- (p.amount * o.platform_fee_rate)::bigint
				- (p.amount * p.processing_fee_rate)::bigint)          AS net_to_organizer,
			o.platform_fee_rate                                          AS platform_fee_rate,
			p.processing_fee_rate                                        AS processing_fee_rate,

			-- Registration
			r.id                                AS registration_id,
			r.registration_number,

			-- Attendee (user preferred, guest fallback)
			COALESCE(u.name,  r.guest_name,  '') AS attendee_name,
			COALESCE(u.email, r.guest_email, '') AS attendee_email,
			COALESCE(u.phone, r.guest_phone, '') AS attendee_phone,

			-- Event
			e.id                                AS event_id,
			COALESCE(e.display_name, e.name, '') AS event_title,
			e.start_date                        AS event_start_date,
			COALESCE(e.image_url, '')           AS event_image_url,

			-- Timing
			p.initiated_at,
			p.completed_at,
			p.created_at
		FROM payments p
		JOIN orders o               ON o.id = p.order_id
		JOIN registrations r        ON r.id = o.registration_id
		JOIN event_registrations er ON er.registration_id = r.id
		JOIN events e               ON e.id = er.event_id
		LEFT JOIN users u           ON u.id = r.user_id
		WHERE ` + whereSQL + `
		ORDER BY ` + orderClause + `
		LIMIT ? OFFSET ?
	`
	dataArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)

	var rows []*paymentdomain.PaymentListRow
	if err := r.db.WithContext(ctx).Raw(dataSQL, dataArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("list payments: %w", err)
	}

	return rows, int(total), nil
}

// buildPaymentOrderClause maps the domain sort fields onto the SQL
// ORDER BY clause. Defaults to created_at DESC when the field is
// unrecognised.
//
// NULLS LAST is used for completed_at because pending payments have
// no completion timestamp and should sort to the bottom.
func buildPaymentOrderClause(sortBy, sortOrder string) string {
	dir := "DESC"
	if strings.EqualFold(sortOrder, "asc") {
		dir = "ASC"
	}
	switch sortBy {
	case "amount":
		return "p.amount " + dir
	case "completed_at":
		return "p.completed_at " + dir + " NULLS LAST"
	case "created_at", "":
		return "p.created_at " + dir
	default:
		return "p.created_at DESC"
	}
}

// AggregateForAccount returns summary statistics for one account,
// using the same filters as ListForAccount (minus pagination).
//
// The fee breakdown is computed in SQL so the totals match what the
// list endpoint shows — no drift between the two.
func (r *PaymentRepository) AggregateForAccount(
	ctx context.Context,
	f paymentdomain.ListPaymentsFilter,
) (*paymentdomain.PaymentStats, error) {
	if f.BilledAccountID == "" {
		return &paymentdomain.PaymentStats{ByStatus: map[string]int64{}}, nil
	}

	where := []string{
		"o.billed_account_id = ?",
		"o.deleted_at IS NULL",
		"r.deleted_at IS NULL",
		"e.deleted_at IS NULL",
	}
	args := []any{f.BilledAccountID}

	if f.EventID != "" {
		where = append(where, "e.id = ?")
		args = append(args, f.EventID)
	}
	if f.DateFrom != nil {
		where = append(where, "p.created_at >= ?")
		args = append(args, *f.DateFrom)
	}
	if f.DateTo != nil {
		where = append(where, "p.created_at <= ?")
		args = append(args, *f.DateTo)
	}

	whereSQL := strings.Join(where, " AND ")

	// Only successful and refunded payments count toward revenue.
	// Pending / failed / expired payments moved no money, so they
	// contribute 0. Filtering in SQL keeps the totals unambiguous.
	sql := `
		SELECT
			COALESCE(SUM(p.amount) FILTER (WHERE p.status IN ('succeeded', 'refunded')), 0)                                          AS total_revenue,
			COALESCE(SUM((p.amount * o.platform_fee_rate)::bigint) FILTER (WHERE p.status IN ('succeeded', 'refunded')), 0)          AS total_platform_fees,
			COALESCE(SUM((p.amount * p.processing_fee_rate)::bigint) FILTER (WHERE p.status IN ('succeeded', 'refunded')), 0)        AS total_processing_fees,
			COALESCE(SUM(p.amount
				- (p.amount * o.platform_fee_rate)::bigint
				- (p.amount * p.processing_fee_rate)::bigint)
				FILTER (WHERE p.status IN ('succeeded', 'refunded')), 0)                                                              AS total_net,
			COUNT(*) FILTER (WHERE p.status IN ('succeeded', 'refunded'))                                                            AS transaction_count,
			COALESCE(MAX(p.currency), 'KES')                                                                                          AS currency,
			COUNT(*) FILTER (WHERE p.status = 'succeeded')                                                                            AS succeeded_count,
			COUNT(*) FILTER (WHERE p.status = 'pending')                                                                              AS pending_count,
			COUNT(*) FILTER (WHERE p.status = 'failed')                                                                               AS failed_count,
			COUNT(*) FILTER (WHERE p.status = 'refunded')                                                                             AS refunded_count,
			COUNT(*) FILTER (WHERE p.status = 'expired')                                                                              AS expired_count
		FROM payments p
		JOIN orders o               ON o.id = p.order_id
		JOIN registrations r        ON r.id = o.registration_id
		JOIN event_registrations er ON er.registration_id = r.id
		JOIN events e               ON e.id = er.event_id
		WHERE ` + whereSQL

	// Anonymous scan target — tags confined to the repo.
	var row struct {
		TotalRevenue        int64  `gorm:"column:total_revenue"`
		TotalPlatformFees   int64  `gorm:"column:total_platform_fees"`
		TotalProcessingFees int64  `gorm:"column:total_processing_fees"`
		TotalNet            int64  `gorm:"column:total_net"`
		TransactionCount    int64  `gorm:"column:transaction_count"`
		Currency            string `gorm:"column:currency"`
		SucceededCount      int64  `gorm:"column:succeeded_count"`
		PendingCount        int64  `gorm:"column:pending_count"`
		FailedCount         int64  `gorm:"column:failed_count"`
		RefundedCount       int64  `gorm:"column:refunded_count"`
		ExpiredCount        int64  `gorm:"column:expired_count"`
	}

	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("aggregate payments: %w", err)
	}

	return &paymentdomain.PaymentStats{
		TotalRevenue:        row.TotalRevenue,
		TotalPlatformFees:   row.TotalPlatformFees,
		TotalProcessingFees: row.TotalProcessingFees,
		TotalNet:            row.TotalNet,
		TransactionCount:    row.TransactionCount,
		Currency:            row.Currency,
		ByStatus: map[string]int64{
			"succeeded": row.SucceededCount,
			"pending":   row.PendingCount,
			"failed":    row.FailedCount,
			"refunded":  row.RefundedCount,
			"expired":   row.ExpiredCount,
		},
	}, nil
}

// ============================================================
// COMPILE-TIME ASSERTION
// ============================================================

var _ paymentdomain.PaymentRepository = (*PaymentRepository)(nil)