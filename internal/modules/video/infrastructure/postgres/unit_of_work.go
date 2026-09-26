// internal/modules/video/infrastructure/postgres/unit_of_work.go

package postgres

import (
	"context"

	"gorm.io/gorm"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// UnitOfWork implements videodomain.UnitOfWork against Postgres.
//
// It runs a callback inside a database transaction. The repositories
// passed to the callback share the same transaction handle, so every
// write through them commits or rolls back together.
type UnitOfWork struct {
	db     *gorm.DB
	cipher videodomain.TokenCipher
}

func NewUnitOfWork(db *gorm.DB, cipher videodomain.TokenCipher) *UnitOfWork {
	return &UnitOfWork{db: db, cipher: cipher}
}

// Do executes fn inside a transaction.
//
// If fn returns an error, the transaction is rolled back and the error
// is propagated. If fn returns nil, the transaction is committed.
func (u *UnitOfWork) Do(
	ctx context.Context,
	fn func(videodomain.Repositories) error,
) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(videodomain.Repositories{
			Connections: NewConnectionRepository(tx, u.cipher),
			OAuthStates: NewOAuthStateRepository(tx),
			Meetings:    NewMeetingRepository(tx),
		})
	})
}

// compile-time assertion
var _ videodomain.UnitOfWork = (*UnitOfWork)(nil)