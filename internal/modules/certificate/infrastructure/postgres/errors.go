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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// The unique index on system_slug was already dropped in
			// migration 157, so this shouldn't fire anymore. But keep
			// the guard for safety if it's ever re-added.
			if pgErr.ConstraintName == "idx_system_templates_unique_slug" {
				return fmt.Errorf("%w: %s", domain.ErrSystemTemplateAlreadyExists, op)
			}
		}
		// Certificate identity / verification_code collisions.
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
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return strings.Contains(err.Error(), "violates foreign key constraint")
}