package postgres

import (
	"context"

	"gorm.io/gorm"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// UnitOfWork implements attendance.UnitOfWork against Postgres.
//
// It runs a callback inside a database transaction. The repositories
// passed to the callback share the same *gorm.DB transaction handle,
// so every write through them commits or rolls back together.
type UnitOfWork struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

// Do executes fn inside a transaction.
//
// If fn returns an error, the transaction is rolled back and the error
// is propagated. If fn returns nil, the transaction is committed.
func (u *UnitOfWork) Do(
	ctx context.Context,
	fn func(attendance.Repositories) error,
) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(attendance.Repositories{
			Attendees:       NewAttendeeRepository(tx),
			Sessions:        NewSessionRepository(tx),
			JoinTokens:      NewJoinTokenRepository(tx),
			Records:         NewAttendanceRecordRepository(tx),
			SessionStatuses: NewAttendeeSessionStatusRepository(tx),
			RollupStatuses:  NewAttendeeRollupStatusRepository(tx),
			Overrides:       NewAttendanceOverrideRepository(tx),
		})
	})
}

// compile-time assertion
var _ attendance.UnitOfWork = (*UnitOfWork)(nil)