
// internal/modules/video/service/service.go

package service

import (
	"context"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// Service is the video module's public API.
//
// Consumers (events module, frontend handlers) depend on this
// interface, not on the concrete implementation.
type Service interface {
	// ============================================================
	// CONNECTION MANAGEMENT
	// ============================================================

	// BeginConnect starts the OAuth flow for the given user and
	// platform. Returns the URL the user should be redirected to and
	// the state that must be persisted (it is — this is informational).
	BeginConnect(ctx context.Context, cmd BeginConnectCommand) (*ConnectResult, error)

	// HandleCallback is called when the platform redirects back to
	// Nuruvent. It validates the state, exchanges the code, and
	// stores the connection. Returns the newly-created connection.
	HandleCallback(ctx context.Context, cmd CallbackCommand) (*ConnectionResult, error)

	// Disconnect revokes and deletes (soft) the user's connection to
	// the given platform. Idempotent.
	Disconnect(ctx context.Context, cmd DisconnectCommand) error

	// ============================================================
	// CONNECTION QUERIES
	// ============================================================

	// GetConnection returns the user's active connection for the
	// given platform.
	//
	// Returns ErrConnectionNotFound if none exists.
	GetConnection(
		ctx context.Context,
		userID string,
		platform videodomain.Platform,
	) (*videodomain.Connection, error)

	// ListConnections returns all of a user's connections (active
	// and revoked), newest first.
	ListConnections(ctx context.Context, userID string) ([]*videodomain.Connection, error)

	// IsConnected reports whether the user has an active connection
	// for the given platform.
	IsConnected(
		ctx context.Context,
		userID string,
		platform videodomain.Platform,
	) (bool, error)

	// ============================================================
	// MEETINGS
	// ============================================================

	// CreateMeeting creates a meeting on the connected host's
	// platform account. Requires an active connection.
	CreateMeeting(ctx context.Context, cmd CreateMeetingCommand) (*videodomain.Meeting, error)

	// DeleteMeeting removes a meeting from the platform. Best-effort
	// at the platform; always removes the local record.
	DeleteMeeting(ctx context.Context, cmd DeleteMeetingCommand) error

	// ============================================================
	// MAINTENANCE
	// ============================================================

	// CleanupExpiredOAuthStates removes oauth_state rows past their
	// expiry window. Called by a background job.
	CleanupExpiredOAuthStates(ctx context.Context) (int, error)
}