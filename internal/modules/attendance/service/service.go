
// internal/modules/attendance/service/service.go

package service

import (
	"context"
	"time"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// Service is the attendance module's public API.
//
// Consumers (registration, courses, admin tools) depend on this
// interface. The concrete implementation lives in the same package.
type Service interface {
	// ============================================================
	// Consumer-facing: register attendees and sessions
	// ============================================================

	// RegisterAttendee registers a new attendee for an external
	// reference. Idempotent: re-registering the same external
	// reference returns the existing attendee.
	RegisterAttendee(ctx context.Context, cmd RegisterAttendeeCommand) (*attendance.Attendee, error)

	// UpsertSession creates or updates a session for an external
	// reference. Idempotent on (external_type, external_id).
	UpsertSession(ctx context.Context, cmd UpsertSessionCommand) (*attendance.Session, error)

	// ============================================================
	// Join tokens
	// ============================================================

	// IssueJoinToken generates a new join token for an attendee in a
	// session.
	//
	// Returns the raw token and its expiry. The raw token is never
	// persisted; only its SHA-256 hash. When the command carries a
	// PublicBaseURL, the result also includes a full join URL.
	IssueJoinToken(
	ctx context.Context,
	cmd IssueJoinTokenCommand,
) (*IssueJoinTokenResult, error)

	// RevokeJoinTokens revokes every active token for a
	// (attendee, session) pair.
	RevokeJoinTokens(ctx context.Context, cmd RevokeJoinTokensCommand) error

	// RedeemJoinToken validates a raw token, records a join, and
	// returns the session to redirect to.
	RedeemJoinToken(ctx context.Context, rawToken string) (*RedeemResult, error)

	// ============================================================
	// Attendance recording
	// ============================================================

	// RecordJoin records a join event for an attendee in a session.
	RecordJoin(ctx context.Context, cmd RecordJoinCommand) (*attendance.AttendanceRecord, error)

	// RecordLeave closes the most recent open record for an attendee
	// in a session.
	RecordLeave(ctx context.Context, cmd RecordLeaveCommand) error

	// ============================================================
	// Webhook ingestion
	// ============================================================

	// IngestWebhook handles a normalized webhook event from a video
	// provider. It resolves the session, matches the participant to
	// an attendee, records the join/leave, and triggers
	// recomputation.
	IngestWebhook(ctx context.Context, provider attendance.SessionProvider, payload []byte, headers map[string]string) error

	// ProcessWebhookEvent handles an already-parsed, already-verified
	// webhook event. The delivery layer parses the raw payload (so it
	// can intercept provider-specific verification handshakes like
	// Zoom's endpoint.url_validation), then hands the normalized
	// event to the service via this method.
	ProcessWebhookEvent(ctx context.Context, provider attendance.SessionProvider, event *attendance.WebhookEvent) error

	// ============================================================
	// Host operations
	// ============================================================

	// ConfirmAttendance marks an attendee as confirmed by a host.
	ConfirmAttendance(ctx context.Context, cmd ConfirmAttendanceCommand) error

	// OverrideAttendance corrects an attendee's derived status.
	// Audit record is written.
	OverrideAttendance(ctx context.Context, cmd OverrideAttendanceCommand) error

	// BulkConfirm confirms several attendees in one call.
	BulkConfirm(ctx context.Context, cmd BulkConfirmCommand) error

	// ============================================================
	// Reporting
	// ============================================================

	// ListSessionAttendance returns all attendee statuses for a
	// session.
	ListSessionAttendance(ctx context.Context, sessionID string) ([]*attendance.AttendeeSessionStatus, error)

	// GetAttendeeSummary returns an attendee's status across every
	// session they're registered for.
	GetAttendeeSummary(ctx context.Context, attendeeID string) (*AttendeeSummary, error)

	// ExportSessionAttendance returns a CSV for a session.
	ExportSessionAttendance(ctx context.Context, sessionID string) ([]byte, error)

	// ============================================================
	// Background jobs
	// ============================================================

	// TransitionSessions advances session lifecycle (scheduled → live
	// → ended) based on the current time.
	TransitionSessions(ctx context.Context, now time.Time) error

	// RecomputeSessionStatuses recomputes every attendee's status for
	// a session, and updates the rollup for each attendee.
	RecomputeSessionStatuses(ctx context.Context, sessionID string) error

	// RecomputeRollup recomputes the roll-up for one attendee and one
	// external reference.
	RecomputeRollup(ctx context.Context, attendeeID string, ref attendance.ExternalRef) error


		// SetAttendeeGoogleMeetID persists a Google Meet user id on an
	// attendee row. Used by the video module to link an unmatched Meet
	// participant to a registered attendee.
	SetAttendeeGoogleMeetID(ctx context.Context, attendeeID, googleMeetUserID string) error


	// RegisterAttendeeForExternal registers an attendee for every session
// under the given external reference. Creates a status row per session
// with derived_status = registered, and issues one join token per
// session.
//
// Returns one AttendeeSessionLink per session. When the command carries
// a PublicBaseURL, each link contains a full join URL. When it does
// not, links carry only the platform meeting code.
//
// Idempotent for status rows: re-registering does not create duplicate
// rows. Tokens are always issued fresh.
RegisterAttendeeForExternal(
	ctx context.Context,
	cmd RegisterAttendeeForExternalCommand,
) (*RegisterAttendeeForExternalResult, error)


		// ListSessionsForExternal returns every session under an external
	// reference. Used by consumers that need to iterate a parent's
	// sessions (e.g. issuing join tokens per session).
	ListSessionsForExternal(ctx context.Context, ref attendance.ExternalRef) ([]*attendance.Session, error)



	




	// GetEventAttendanceSummary returns a per-session and total
	// attendance summary for every session under an event.
	GetEventAttendanceSummary(
		ctx context.Context,
		cmd GetEventAttendanceSummaryCommand,
	) (*EventAttendanceSummary, error)



	// RecordExternalParticipant records a join or leave observed on an
	// external platform (Google Meet, Teams, etc.).
	//
	// The video module calls this after fetching participants from the
	// platform's API. It is the counterpart to ProcessWebhookEvent for
	// platforms that don't push events.
	RecordExternalParticipant(
		ctx context.Context,
		cmd RecordExternalParticipantCommand,
	) error
}