// internal/modules/payment/infrastructure/postgres/unit_of_work.go

package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// UnitOfWork implements paymentdomain.UnitOfWork against Postgres.
//
// It runs a callback inside a database transaction. The repositories
// passed to the callback share the same *gorm.DB transaction handle,
// so every write through them commits or rolls back together.
type UnitOfWork struct {
	db *gorm.DB
}

// NewUnitOfWork constructs a UnitOfWork.
func NewUnitOfWork(db *gorm.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

// Do executes fn inside a transaction.
//
// If fn returns an error, the transaction is rolled back and the error
// is propagated. If fn returns nil, the transaction is committed.
func (u *UnitOfWork) Do(
	ctx context.Context,
	fn func(paymentdomain.Repositories) error,
) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(paymentdomain.Repositories{
			Orders:   NewOrderRepository(tx),
			Payments: NewPaymentRepository(tx),
			Refunds:  NewRefundRepository(tx),
			Webhooks: NewWebhookEventRepository(tx),
		})
	})
}

// Compile-time assertion.
var _ paymentdomain.UnitOfWork = (*UnitOfWork)(nil)