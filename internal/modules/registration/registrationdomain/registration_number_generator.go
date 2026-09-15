// internal/modules/registration/registrationdomain/registration_number_generator.go

package registrationdomain

import "context"

// RegistrationNumberGenerator produces unique, human-readable registration
// numbers in a per-year sequential format (e.g. "REG-2026-000123").
//
// The domain defines this port so the service layer doesn't depend on
// Postgres or any specific implementation. The Postgres-backed
// implementation lives in infrastructure/postgres.
type RegistrationNumberGenerator interface {
	// Next returns the next registration number. Implementations must be
	// safe for concurrent use and must not return duplicate values.
	Next(ctx context.Context) (string, error)
}