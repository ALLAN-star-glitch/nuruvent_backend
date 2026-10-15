// internal/modules/certificate/infrastructure/postgres/errors.go

package postgres

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

// translateError converts a GORM/Postgres error into a domain error
// where appropriate. Preserves the underlying error with %w.
func translateError(err error, op string) error {
	if err == nil {
		return nil
	}

	if isUniqueViolation(err) {
		// Both identity-partial-indexes and verification_code collisions
		// surface here. We can't distinguish which index from the driver
		// without parsing the constraint name — the caller (application
		// service) treats them as ErrAlreadyIssued to remain idempotent.
		return fmt.Errorf("%w: %s", domain.ErrAlreadyIssued, op)
	}

	if isForeignKeyViolation(err) {
		return fmt.Errorf("%s: foreign key violation: %w", op, err)
	}

	return fmt.Errorf("%s: %w", op, err)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	// Fallback for drivers that don't expose pgconn.PgError
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return strings.Contains(err.Error(), "violates foreign key constraint")
}