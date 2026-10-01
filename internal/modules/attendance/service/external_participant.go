// internal/modules/attendance/service/external_participant.go

package service

import (
	"context"
	"time"
)

// ExternalParticipantEventType identifies whether an external
// participant event is a join or a leave.
//
// String values match the constants used elsewhere in the module so
// adapters can safely cast between them.
type ExternalParticipantEventType string

const (
	ExternalParticipantJoined ExternalParticipantEventType = "joined"
	ExternalParticipantLeft   ExternalParticipantEventType = "left"
)

// IsValid reports whether the event type is one we recognize.
func (t ExternalParticipantEventType) IsValid() bool {
	return t == ExternalParticipantJoined || t == ExternalParticipantLeft
}

// RecordExternalParticipantCommand is the input to
// RecordExternalParticipant.
//
// It describes a participant observed on an external platform. The
// attendance service resolves the session, matches the participant
// to a registered attendee, and records the event.
//
// Matching strategy:
//
//  1. Google Meet user id — ExternalUserID, when the attendee row
//     has a previously linked google_meet_user_id.
//  2. Display name — normalized, case-insensitive.
//
// ParticipantEmail is deliberately absent: the Google Meet API does
// not expose one for external participants.
type RecordExternalParticipantCommand struct {
	// Provider identifies the source platform, e.g. "google_meet".
	Provider string

	// MeetingCode is the platform's external meeting identifier,
	// e.g. "spaces/efoBh0pzi5IB". Used to resolve the session via
	// (provider, provider_meeting_id).
	MeetingCode string

	// ParticipantName is the display name as shown in the meeting.
	// Used for attendee matching when no identity match is found.
	ParticipantName string

	// ExternalUserID is the platform's user identifier when
	// available, e.g. Google's "users/123456". Empty for anonymous
	// participants.
	//
	// For Google Meet, this is checked first against the attendee's
	// stored google_meet_user_id. If it matches, the participant is
	// resolved to that attendee without touching display names. If
	// there's no identity match, the name fallback runs and — on
	// success — ExternalUserID is persisted on the matched attendee
	// row so the next poll takes the identity path.
	ExternalUserID string

	// OccurredAt is when the event happened at the provider.
	OccurredAt time.Time

	// EventType is join or leave.
	EventType ExternalParticipantEventType
}

// ParticipantRecorder is the attendance module's contract for
// recording external participants and linking platform identities.
//
// Implemented by the attendance service. Consumed by cross-module
// adapters (see internal/app/adapters) so modules like video can
// hand off observed participants without importing attendance
// directly.
type ParticipantRecorder interface {
	// RecordExternalParticipant records a join or leave for a
	// platform participant.
	//
	// Returns ErrParticipantUnmatched when the participant cannot be
	// resolved to a registered attendee. Callers treat that as a
	// soft signal — see the video module's ParticipantRecorder for
	// the port-level counterpart.
	RecordExternalParticipant(
		ctx context.Context,
		cmd RecordExternalParticipantCommand,
	) error

	// SetAttendeeGoogleMeetID persists a Google Meet user id on an
	// attendee row. Used by the participant-linking flow: the host
	// picks an attendee from the roster, the video handler calls
	// this to store the mapping, then re-polls the meeting.
	SetAttendeeGoogleMeetID(
		ctx context.Context,
		attendeeID, googleMeetUserID string,
	) error
}