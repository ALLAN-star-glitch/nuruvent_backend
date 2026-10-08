package attendancedomain

import "context"

// RegistrationLookup resolves a registration ID to the user behind
// it. Used by the attendance module when a guest redeems a join
// token, so the join handler can issue a session cookie for them.
//
// The attendance module does not import the registration module —
// this port is implemented by an adapter in the app layer.
type RegistrationLookup interface {
	// GetUserIDByRegistrationID returns the user_id column of the
	// registration with the given ID. Returns an empty string and no
	// error when the registration is not found or has no user.
	GetUserIDByRegistrationID(ctx context.Context, registrationID string) (string, error)
}