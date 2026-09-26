// internal/modules/video/infrastructure/postgres/errors.go

package postgres

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// translateError converts a GORM/Postgres error into a domain error
// where appropriate. The notFoundSentinel argument selects which
// domain "not found" error to use when a missing record is detected.
func translateError(err error, op string, notFoundSentinel error) error {
	if err == nil {
		return nil
	}
	if isUniqueViolation(err) {
		return fmt.Errorf("%w: %s", videodomain.ErrInvalidConnection, op)
	}
	if isForeignKeyViolation(err) {
		return fmt.Errorf("%w: %s (foreign key)", notFoundSentinel, op)
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