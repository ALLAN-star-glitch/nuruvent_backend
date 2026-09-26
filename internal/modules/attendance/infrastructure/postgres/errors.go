package postgres

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// translateError converts a GORM/Postgres error into a domain error
// where appropriate. Preserves the underlying error with %w.
//
// The `notFoundSentinel` argument selects the domain sentinel to use
// when a record-not-found or FK violation is detected.
func translateError(err error, op string, notFoundSentinel error) error {
	if err == nil {
		return nil
	}

	if isUniqueViolation(err) {
		// The sentinel depends on the table — callers pass the right one.
		return fmt.Errorf("%w: %s", notFoundSentinel, op)
	}
	if isForeignKeyViolation(err) {
		return fmt.Errorf("%w: %s (foreign key)", notFoundSentinel, op)
	}
	return fmt.Errorf("%s: %w", op, err)
}

// translateDuplicate converts a unique violation into a specific
// duplicate sentinel, preserving other errors through translateError.
func translateDuplicate(err error, op string, duplicateSentinel error) error {
	if err == nil {
		return nil
	}
	if isUniqueViolation(err) {
		return fmt.Errorf("%w: %s", duplicateSentinel, op)
	}
	return fmt.Errorf("%s: %w", op, err)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}

// compile-time assertion
var _ = attendance.ErrAttendeeNotFound