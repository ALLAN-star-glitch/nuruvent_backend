// internal/modules/registration/registrationdomain/status.go

package registrationdomain

import (
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// Status is the lifecycle state of a registration.
// Aliased to the shared types.RegistrationStatus so the DB reference table
// and the domain use the same value type.
type Status = types.RegistrationStatus

const (
	StatusPending   Status = types.RegistrationStatusPending
	StatusConfirmed Status = types.RegistrationStatusConfirmed
	StatusCancelled Status = types.RegistrationStatusCancelled
	StatusRefunded  Status = types.RegistrationStatusRefunded
	StatusExpired   Status = types.RegistrationStatusExpired
	StatusAttended  Status = types.RegistrationStatusAttended
)

// ============================================================
// DOMAIN BEHAVIOUR — methods that don't live on types.RegistrationStatus
// ============================================================

// IsTerminal reports whether the status represents a final state.
func isTerminal(s Status) bool {
	switch s {
	case StatusCancelled, StatusRefunded, StatusExpired:
		return true
	}
	return false
}

// IsActive reports whether the registration counts toward capacity.
func isActive(s Status) bool {
	return s == StatusConfirmed || s == StatusPending
}

// CanTransitionTo validates whether a status change is allowed.
func CanTransitionTo(from, to Status) bool {
	switch from {
	case StatusPending:
		return to == StatusConfirmed ||
			to == StatusExpired ||
			to == StatusCancelled
	case StatusConfirmed:
		return to == StatusCancelled ||
			to == StatusRefunded ||
			to == StatusAttended
	case StatusAttended:
		return to == StatusRefunded
	case StatusCancelled, StatusRefunded, StatusExpired:
		return false
	}
	return false
}

// ParseStatus converts a string into a Status. Kept for compatibility
// with repository code that reads raw strings from the DB.
func ParseStatus(v string) (Status, error) {
	s, ok := types.ParseRegistrationStatus(v)
	if !ok {
		return "", fmt.Errorf("invalid registration status: %q", v)
	}
	return s, nil
}