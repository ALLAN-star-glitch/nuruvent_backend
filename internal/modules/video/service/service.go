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

	BeginConnect(ctx context.Context, cmd BeginConnectCommand) (*ConnectResult, error)
	HandleCallback(ctx context.Context, cmd CallbackCommand) (*ConnectionResult, error)
	Disconnect(ctx context.Context, cmd DisconnectCommand) error

	// ============================================================
	// CONNECTION QUERIES
	// ============================================================

	GetConnection(
		ctx context.Context,
		userID string,
		platform videodomain.Platform,
	) (*videodomain.Connection, error)

	ListConnections(ctx context.Context, userID string) ([]*videodomain.Connection, error)

	IsConnected(
		ctx context.Context,
		userID string,
		platform videodomain.Platform,
	) (bool, error)

	// ============================================================
	// MEETINGS
	// ============================================================

	CreateMeeting(ctx context.Context, cmd CreateMeetingCommand) (*videodomain.Meeting, error)

	UpdateMeeting(ctx context.Context, cmd UpdateMeetingCommand) (*videodomain.Meeting, error)

	DeleteMeeting(ctx context.Context, cmd DeleteMeetingCommand) error


		// FindMeetingIDByProviderMeeting returns the internal ID of the
	// video meeting identified by (platform, provider meeting code).
	//
	// Returns ("", nil) if no meeting matches — that is not an error.
	// Used by the attendance module (via an adapter) to link attendance
	// sessions to their underlying video meetings.
	FindMeetingIDByProviderMeeting(
		ctx context.Context,
		provider string,
		providerMeetingID string,
	) (string, error)

	// ============================================================
	// ATTENDANCE
	// ============================================================

	// FetchGoogleMeetAttendance polls Google Meet for conference
	// records and participants, and dispatches events to the
	// attendance module.
	//
	// Only supported for google_meet meetings. Zoom attendance
	// arrives via webhook and does not use this path.
	FetchGoogleMeetAttendance(
		ctx context.Context,
		cmd FetchGoogleMeetAttendanceCommand,
	) (*FetchGoogleMeetAttendanceResult, error)

	// ============================================================
	// MEETING SDK
	// ============================================================

	GenerateMeetingSignature(ctx context.Context, cmd GenerateMeetingSignatureCommand) (*videodomain.MeetingSignature, error)

	FetchMeetingZAK(ctx context.Context, cmd FetchMeetingZAKCommand) (*videodomain.MeetingZAK, error)

	GetMeetingJoinInfo(ctx context.Context, cmd GetMeetingJoinInfoCommand) (*MeetingJoinInfo, error)

	// ============================================================
	// MAINTENANCE
	// ============================================================

	CleanupExpiredOAuthStates(ctx context.Context) (int, error)


		// LinkParticipant binds a Meet participant's Google user id to a
	// registered attendee, then re-polls the meeting so the newly-
	// linked identity is picked up and the join is recorded.
	LinkParticipant(
		ctx context.Context,
		cmd LinkParticipantCommand,
	) (*FetchGoogleMeetAttendanceResult, error)

	// GetUnmatchedParticipants polls the meeting and returns any
	// participants that couldn't be resolved to a registered attendee.
	GetUnmatchedParticipants(
		ctx context.Context,
		cmd GetUnmatchedParticipantsCommand,
	) ([]UnmatchedParticipant, error)

		// AutoFetchGoogleMeetAttendance runs a background poller that
	// fetches Google Meet attendance for every meeting whose
	// scheduled window is currently active.
	//
	// Returns when ctx is cancelled. Idempotent — repeated fetches
	// for the same meeting are deduplicated inside
	// FetchGoogleMeetAttendance.
	//
	// Started once at boot, runs for the process lifetime.
	AutoFetchGoogleMeetAttendance(ctx context.Context) error
}