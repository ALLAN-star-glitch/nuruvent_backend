package service

import (
	"context"
	"errors"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// FindMeetingIDByProviderMeeting implements Service.
//
// It reuses the existing FindByExternalID lookup, which is keyed by
// (platform, external_id) — the same pair the attendance module
// stores as (provider, provider_meeting_id) on a session.
func (s *videoService) FindMeetingIDByProviderMeeting(
	ctx context.Context,
	provider string,
	providerMeetingID string,
) (string, error) {
	provider = strings.TrimSpace(provider)
	providerMeetingID = strings.TrimSpace(providerMeetingID)
	if provider == "" || providerMeetingID == "" {
		return "", nil
	}

	var meetingID string

	err := s.deps.UnitOfWork.Do(ctx, func(repos videodomain.Repositories) error {
		meeting, txErr := repos.Meetings.FindByExternalID(
			ctx,
			videodomain.Platform(provider),
			providerMeetingID,
		)
		if txErr != nil {
			if errors.Is(txErr, videodomain.ErrMeetingNotFound) {
				// Not found is not an error — many sessions
				// legitimately have no linked video meeting.
				return nil
			}
			return txErr
		}
		meetingID = meeting.ID
		return nil
	})
	if err != nil {
		return "", err
	}

	return meetingID, nil
}