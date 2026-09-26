// internal/modules/video/videodomain/repository.go

package videodomain

import (
	"context"
	"time"
)

// ============================================================
// CONNECTIONS
// ============================================================

// ConnectionRepository persists host connections to video platforms.
type ConnectionRepository interface {
	// Create inserts a new connection.
	Create(ctx context.Context, conn *Connection) error

	// Update persists changes to an existing connection. Used for
	// token refresh and for marking revoked.
	Update(ctx context.Context, conn *Connection) error

	// FindByID returns a connection by its Nuruvent ID.
	FindByID(ctx context.Context, id string) (*Connection, error)

	// FindActiveByUserAndPlatform returns the single active
	// connection for a user on a platform, if one exists.
	//
	// Returns ErrConnectionNotFound if none exists.
	FindActiveByUserAndPlatform(
		ctx context.Context,
		userID string,
		platform Platform,
	) (*Connection, error)

	// ListByUser returns all connections (active and revoked) for a
	// user, ordered newest first.
	ListByUser(ctx context.Context, userID string) ([]*Connection, error)

	// Revoke marks a connection as revoked. Idempotent.
	Revoke(ctx context.Context, id string, now time.Time) error
}

// ============================================================
// OAUTH STATES
// ============================================================

// OAuthStateRepository persists single-use OAuth state records.
type OAuthStateRepository interface {
	// Create inserts a new state.
	Create(ctx context.Context, state *OAuthState) error

	// Consume atomically fetches and marks a state as consumed.
	//
	// Returns ErrOAuthStateNotFound if no record exists for the
	// state. Returns ErrOAuthStateConsumed if the state exists but
	// was already consumed. Returns ErrOAuthStateExpired if the
	// state exists and is unconsumed but past its expiry.
	//
	// Implementations must make the check-and-mark atomic — two
	// concurrent callbacks with the same state must not both succeed.
	Consume(ctx context.Context, state string, now time.Time) (*OAuthState, error)

	// DeleteExpired removes states past their expiry window. Called
	// by a background job. Returns the number deleted.
	DeleteExpired(ctx context.Context, before time.Time) (int, error)
}

// ============================================================
// MEETINGS
// ============================================================

// MeetingRepository persists meetings created via providers.
//
// Meetings are stored so the system can look them up later — for
// deletion, for updating after a schedule change, or for reconciling
// provider webhooks.
type MeetingRepository interface {
	// Create inserts a new meeting.
	Create(ctx context.Context, m *Meeting) error

	// Update persists changes to an existing meeting.
	Update(ctx context.Context, m *Meeting) error

	// FindByID returns a meeting by its Nuruvent ID.
	FindByID(ctx context.Context, id string) (*Meeting, error)

	// FindByExternalID returns a meeting by its platform-side
	// identifier. Used by reconciliation jobs and by any handler
	// that receives a webhook and needs to find the corresponding
	// Nuruvent meeting.
	FindByExternalID(
		ctx context.Context,
		platform Platform,
		externalID string,
	) (*Meeting, error)

	// Delete removes a meeting record. Called when a schedule is
	// removed and the platform meeting has been deleted.
	Delete(ctx context.Context, id string) error
}

// ============================================================
// UNIT OF WORK
// ============================================================

// Repositories groups the module's repositories for use inside a
// transaction.
type Repositories struct {
	Connections ConnectionRepository
	OAuthStates OAuthStateRepository
	Meetings    MeetingRepository
}

// UnitOfWork executes a function inside a database transaction. The
// repositories passed to the callback share the same transaction
// handle.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(Repositories) error) error
}