// internal/modules/video/videodomain/user_lookup.go

package videodomain

import "context"

// UserLookup is the driven port for resolving Nuruvent user display
// data. Implemented by an adapter in internal/app/adapters/video/
// that bridges to the auth or account module.
//
// This is deliberately narrow. The video module needs a display name
// to stamp onto the Zoom Meeting SDK join call — nothing else. Adding
// more methods here should be justified against that scope.
type UserLookup interface {
	// GetDisplayName returns the human-readable name for a user.
	// Falls back to the local part of the email if no display name
	// is set. Returns an empty string (with nil error) when the user
	// doesn't exist — callers should then fall back to a generic
	// label.
	GetDisplayName(ctx context.Context, userID string) (string, error)



	GetEmail(ctx context.Context, userID string) (string, error)
}