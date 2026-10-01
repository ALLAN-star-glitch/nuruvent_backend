// internal/modules/attendance/service/dependencies.go

package service

import (
	"context"
	"time"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	id "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/id"
)

// ============================================================
// CROSS-CUTTING PORTS
// ============================================================

type Clock interface {
	Now() time.Time
}

type TokenGenerator interface {
	Generate() (rawToken string, tokenHash string, err error)
	Hash(rawToken string) string
}

// ============================================================
// PUBLISHER
// ============================================================

type StatusPublisher interface {
	SessionStatusChanged(ctx context.Context, change SessionStatusChange) error
	RollupStatusChanged(ctx context.Context, change RollupStatusChange) error
}

type SessionStatusChange struct {
	AttendeeID       string
	AttendeeExternal attendance.ExternalRef
	SessionID        string
	SessionExternal  attendance.ExternalRef
	OldStatus        attendance.AttendanceStatus
	NewStatus        attendance.AttendanceStatus
	OccurredAt       time.Time
}

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

	VideoMeetings VideoMeetingIDResolver

	// AppConfig carries the frontend base URL used to build absolute
	// redirect URLs on join-token redemption. See buildJoinRedirect.
	AppConfig config.AppConfig
}

// URLValidator is an optional interface a ProviderAdapter may
// implement to handle provider-specific endpoint verification.
type URLValidator interface {
	IsURLValidationError(err error) bool
	HandleURLValidation(err error) map[string]string
}