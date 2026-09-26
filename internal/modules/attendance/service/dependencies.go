// internal/modules/attendance/service/dependencies.go

package service

import (
	"context"
	"time"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	id "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/id"
)

// ============================================================
// CROSS-CUTTING PORTS
// ============================================================

// Clock supplies the current time. Injected so the service is
// deterministic under test.
type Clock interface {
	Now() time.Time
}

// TokenGenerator produces raw join tokens and their hashes.
//
//   - Generate returns (rawToken, tokenHash). The caller stores the
//     hash and hands the raw token to the user.
//   - Hash computes the hash of a raw token for lookups at redeem
//     time.
type TokenGenerator interface {
	Generate() (rawToken string, tokenHash string, err error)
	Hash(rawToken string) string
}

// ============================================================
// PUBLISHER
// ============================================================

// StatusPublisher is notified when attendance statuses change.
type StatusPublisher interface {
	SessionStatusChanged(ctx context.Context, change SessionStatusChange) error
	RollupStatusChanged(ctx context.Context, change RollupStatusChange) error
}

// SessionStatusChange describes a change to one (attendee, session)
// status.
type SessionStatusChange struct {
	AttendeeID       string
	AttendeeExternal attendance.ExternalRef
	SessionID        string
	SessionExternal  attendance.ExternalRef
	OldStatus        attendance.AttendanceStatus
	NewStatus        attendance.AttendanceStatus
	OccurredAt       time.Time
}

// RollupStatusChange describes a change to an attendee's roll-up
// status for one external entity.
type RollupStatusChange struct {
	AttendeeID string
	External   attendance.ExternalRef
	OldStatus  attendance.AttendanceStatus
	NewStatus  attendance.AttendanceStatus
	OccurredAt time.Time
}

// ============================================================
// PROVIDER ADAPTER
// ============================================================

// ProviderAdapter normalizes webhooks from one video platform.
type ProviderAdapter interface {
	Provider() attendance.SessionProvider

	ParseWebhook(
		ctx context.Context,
		payload []byte,
		headers map[string]string,
	) (*attendance.WebhookEvent, error)
}

// ============================================================
// DEPENDENCIES
// ============================================================

// Dependencies is the full set of ports the service needs.
type Dependencies struct {
	// Persistence
	UnitOfWork attendance.UnitOfWork

	// Cross-cutting
	Clock          Clock
	IDs            id.Generator
	TokenGenerator TokenGenerator

	// Publishing
	Publisher StatusPublisher

	// Provider adapters, keyed by provider name.
	Providers map[attendance.SessionProvider]ProviderAdapter

	// Policy
	DerivationPolicy attendance.DerivationPolicy

	// Default join token grace period. Applied when a command's Grace
	// is zero.
	JoinTokenGrace time.Duration
}

// internal/modules/attendance/service/dependencies.go

// URLValidator is an optional interface a ProviderAdapter may
// implement to handle provider-specific endpoint verification.
type URLValidator interface {
	IsURLValidationError(err error) bool
	HandleURLValidation(err error) map[string]string
}