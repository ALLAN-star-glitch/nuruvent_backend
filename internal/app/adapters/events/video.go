// internal/app/adapters/events/video.go

package events

import (
	"context"
	"strings"


	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
	videoservice "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/service"
	eventsdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// VideoAdapter bridges events domain's VideoMeetingCreator port to the
// video module's Service. Lives in internal/app/adapters so neither
// module imports the other.
type VideoAdapter struct {
	svc videoservice.Service
}

// NewVideoAdapter constructs the adapter.
func NewVideoAdapter(svc videoservice.Service) *VideoAdapter {
	return &VideoAdapter{svc: svc}
}

// IsConnected reports whether the user has an active connection for
// the given platform slug.
func (a *VideoAdapter) IsConnected(
	ctx context.Context,
	userID string,
	platform string,
) (bool, error) {
	p := videodomain.Platform(strings.ToLower(platform))
	if !p.IsValid() {
		return false, nil // unknown platform → not connected, no error
	}
	return a.svc.IsConnected(ctx, userID, p)
}

// CreateMeeting creates a meeting on the host's connected account.
func (a *VideoAdapter) CreateMeeting(
	ctx context.Context,
	req eventsdomain.VideoMeetingRequest,
) (*eventsdomain.VideoMeetingResult, error) {
	platform := videodomain.Platform(strings.ToLower(req.Platform))

	meeting, err := a.svc.CreateMeeting(ctx, videoservice.CreateMeetingCommand{
		UserID:   req.UserID,
		Platform: platform,
		Spec: videodomain.MeetingSpec{
			Topic:      req.Topic,
			StartTime:  req.StartTime,
			Duration:   req.Duration,
			Timezone:   req.Timezone,
			Agenda:     req.Agenda,
			HostUserID: req.UserID,
		},
	})
	if err != nil {
		return nil, err
	}

	return &eventsdomain.VideoMeetingResult{
		MeetingID: meeting.ID,
		JoinURL:   meeting.JoinURL,
		StartURL:  meeting.StartURL,
	}, nil
}

// DeleteMeeting removes a meeting from the platform.
func (a *VideoAdapter) DeleteMeeting(
	ctx context.Context,
	req eventsdomain.VideoMeetingDeleteRequest,
) error {
	return a.svc.DeleteMeeting(ctx, videoservice.DeleteMeetingCommand{
		MeetingID: req.MeetingID,
		UserID:    req.UserID,
	})
}