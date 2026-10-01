package attendance

import (
	"context"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	attendanceservice "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// videoMeetingIDResolver bridges attendance's VideoMeetingIDResolver
// port directly to the video module's MeetingRepository.
//
// We depend on the repository (not video.Service) deliberately:
// video.Service already depends on attendance.Service via the
// ParticipantRecorder port, and injecting the service here would
// create a wiring cycle.
//
// This is a read-only lookup, so it does not participate in the
// video module's UnitOfWork.
type videoMeetingIDResolver struct {
	meetings videodomain.MeetingRepository
}

// NewVideoMeetingIDResolver returns an attendance-service port
// implemented against the video module's meeting repository.
func NewVideoMeetingIDResolver(
	meetings videodomain.MeetingRepository,
) attendanceservice.VideoMeetingIDResolver {
	return &videoMeetingIDResolver{meetings: meetings}
}

func (a *videoMeetingIDResolver) ResolveMeetingID(
	ctx context.Context,
	provider attendance.SessionProvider,
	providerMeetingID string,
) (string, error) {
	if providerMeetingID == "" {
		return "", nil
	}

	id, err := a.meetings.FindIDByProviderMeeting(
		ctx,
		string(provider),
		providerMeetingID,
	)
	if err != nil {
		// "not found" is already returned as ("", nil) by the repo,
		// so any error here is an infrastructure fault. Still, don't
		// fail the whole summary over one bad lookup — let the caller
		// decide. Return the error and let get_event_summary swallow
		// it (it currently does).
		return "", err
	}
	return id, nil
}