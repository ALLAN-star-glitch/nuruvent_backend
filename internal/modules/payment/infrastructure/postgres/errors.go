// internal/modules/payment/infrastructure/postgres/errors.go

package postgres

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// Postgres SQLSTATE codes we care about.
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	sqlStateUniqueViolation     = "23505"
	sqlStateForeignKeyViolation = "23503"
	sqlStateNotNullViolation    = "23502"
	sqlStateCheckViolation      = "23514"
)

// isUniqueViolation reports whether err is a Postgres unique
// constraint violation. Used to translate duplicate inserts (payments
// by idempotency key, webhook events by provider event ID) into domain
// errors.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == sqlStateUniqueViolation
	}

	// Fallback for cases where GORM has already stringified the error.
	return strings.Contains(err.Error(), "duplicate key value")
}

// isForeignKeyViolation reports whether err is a Postgres foreign key
// violation. Used to translate "referenced row missing" into a domain
// error.
func isForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == sqlStateForeignKeyViolation
	}

	return strings.Contains(err.Error(), "violates foreign key constraint")
}

// isNotNullViolation reports whether err is a Postgres NOT NULL
// violation. Usually a bug — a required field was left empty.
func isNotNullViolation(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == sqlStateNotNullViolation
	}

	return strings.Contains(err.Error(), "null value in column")
}

// isCheckViolation reports whether err is a Postgres CHECK constraint
// violation. Means the domain allowed something the DB rejects — a
// mismatch between domain rules and schema rules.
func isCheckViolation(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == sqlStateCheckViolation
	}

	return strings.Contains(err.Error(), "violates check constraint")
}