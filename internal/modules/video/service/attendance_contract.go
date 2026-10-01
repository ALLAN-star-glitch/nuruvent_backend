// internal/modules/video/service/attendance_contract.go

package service

import (
	"context"
	"errors"
	"time"
)

// ErrParticipantUnmatched is returned by ParticipantRecorder when a
// participant cannot be resolved to a registered attendee for the
// session. The video module treats this as a soft signal — the
// participant is collected into the poll result so a host can link
// them to an attendee later, not treated as a hard failure.
//
// The attendance module has its own sentinel for the same condition;
// the adapter (internal/app/adapters/attendance/to_video.go) is
// responsible for translating between them. This type exists so the
// video module never has to import the attendance domain.
var ErrParticipantUnmatched = errors.New("participant unmatched")

// ParticipantEventType identifies whether a participant event is a
// join or a leave.
type ParticipantEventType string

const (
	EventTypeJoined ParticipantEventType = "joined"
	EventTypeLeft   ParticipantEventType = "left"
)

// RecordParticipantCommand is the video module's request to the
// attendance module to record a participant's join or leave.
type RecordParticipantCommand struct {
	Provider        string
	MeetingCode     string
	ParticipantName string
	ExternalUserID  string
	OccurredAt      time.Time
	EventType       ParticipantEventType
}

// ParticipantRecorder is the video module's view of the attendance
// service.
type ParticipantRecorder interface {
	// RecordExternalParticipant records a participant's join or
	// leave.
	//
	// Returns ErrParticipantUnmatched when the participant cannot be
	// resolved to a registered attendee. Callers treat that as a
	// soft signal, not a failure.
	RecordExternalParticipant(
		ctx context.Context,
		cmd RecordParticipantCommand,
	) error

	// SetAttendeeGoogleMeetID persists a Google Meet user id on an
	// attendee row so subsequent polls match by identity instead of
	// by display name. Used by the participant-linking flow.
	SetAttendeeGoogleMeetID(
		ctx context.Context,
		attendeeID, googleMeetUserID string,
	) error
}