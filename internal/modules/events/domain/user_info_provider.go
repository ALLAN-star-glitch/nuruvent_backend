// internal/modules/events/domain/user_info_provider.go

package domain

import "context"

// ============================================================
// OUTBOUND PORT: UserInfoProvider
// ============================================================
//
// UserInfoProvider exposes read-only user projections to the events module.
// It is used to populate the `creator` field on event responses (when the
// caller has permission to see creator info).
//
// The concrete implementation is provided by the account module via an
// adapter in the composition layer. The events module does not import the
// account module directly.
//
// Contract:
//   - No permission checks are performed here. Callers must verify access
//     (e.g. canViewCreatorInfo) before invoking.
//   - Returns (nil, nil) if the user does not exist. Callers treat this as
//     "no creator info available" rather than an error.

type UserInfoProvider interface {
	// GetUserByID returns basic user info.
	// Returns (nil, nil) if the user does not exist.
	GetUserByID(ctx context.Context, userID string) (*UserInfo, error)

	// GetUserByIDWithDetails is a richer variant. Currently equivalent to
	// GetUserByID, but reserved for future expansion (avatar, bio, etc.)
	// without changing the call site.
	GetUserByIDWithDetails(ctx context.Context, userID string) (*UserInfo, error)

	// GetUsersByIDs returns user info for multiple IDs in one call.
	// Missing IDs are silently skipped (not returned as nil entries).
	GetUsersByIDs(ctx context.Context, userIDs []string) ([]*UserInfo, error)
}

// UserInfo is the projection of user data used for creator display.
//
// It is intentionally minimal — the events module does not need the full
// user entity. If you add fields here, update the adapter to map them.
type UserInfo struct {
	ID          string
	Name        string
	DisplayName string
	Email       string
	Phone       string
	AvatarURL   string
}