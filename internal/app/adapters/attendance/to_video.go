// internal/app/adapters/attendance/to_video.go

package attendance

import (
	"context"
	"errors"

	attendanceDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	attendanceservice "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
	videoservice "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/service"
)

// AttendanceToVideo adapts the attendance module's
// ParticipantRecorder to the video module's ParticipantRecorder.
//
// Both interfaces describe the same capability — "record an
// external participant's join or leave" — but each module declares
// its own so neither has to import the other.
//
// This adapter is the only place that imports both sides. It is the
// translation layer for both data (command structs) and errors
// (sentinel errors), so the two modules stay decoupled.
type AttendanceToVideo struct {
	recorder attendanceservice.ParticipantRecorder
}

// NewAttendanceToVideo wires the adapter.
func NewAttendanceToVideo(recorder attendanceservice.ParticipantRecorder) *AttendanceToVideo {
	return &AttendanceToVideo{recorder: recorder}
}

// RecordExternalParticipant satisfies
// videoservice.ParticipantRecorder.
//
// It translates the video module's command into the attendance
// module's command and delegates. On the way back, it translates
// the attendance-side "unmatched participant" sentinel into the
// video-side one so callers can react without importing the
// attendance domain.
func (a *AttendanceToVideo) RecordExternalParticipant(
	ctx context.Context,
	cmd videoservice.RecordParticipantCommand,
) error {
	err := a.recorder.RecordExternalParticipant(ctx, attendanceservice.RecordExternalParticipantCommand{
		Provider:        cmd.Provider,
		MeetingCode:     cmd.MeetingCode,
		ParticipantName: cmd.ParticipantName,
		ExternalUserID:  cmd.ExternalUserID,
		OccurredAt:      cmd.OccurredAt,
		EventType: attendanceservice.ExternalParticipantEventType(
			cmd.EventType,
		),
	})
	if err != nil {
		if errors.Is(err, attendanceDomain.ErrParticipantUnmatched) {
			return videoservice.ErrParticipantUnmatched
		}
		return err
	}
	return nil
}

// SetAttendeeGoogleMeetID satisfies videoservice.ParticipantRecorder.
//
// Persists a Google Meet user id on an attendee row so subsequent
// polls match by identity instead of by display name.
func (a *AttendanceToVideo) SetAttendeeGoogleMeetID(
	ctx context.Context,
	attendeeID, googleMeetUserID string,
) error {
	return a.recorder.SetAttendeeGoogleMeetID(ctx, attendeeID, googleMeetUserID)
}