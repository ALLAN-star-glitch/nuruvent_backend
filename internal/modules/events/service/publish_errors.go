package service

import (
	"fmt"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// translatePublishError converts a validation error into a user-facing
// message that explains *why* the publish failed and what the user can
// do about it.
//
// Each branch wraps a domain sentinel so the HTTP classifier can
// recognize the specific case and surface the right user message. The
// human-readable text after the sentinel is what the user sees.
//
// validationErr is the error from event.ValidateForPublish().
// meetingErr is the error from attachVideoMeetings(), if any.
func (s *eventService) translatePublishError(validationErr, meetingErr error) error {
	if meetingErr == nil {
		return fmt.Errorf("cannot publish event: %w", validationErr)
	}

	msg := meetingErr.Error()

	switch {
	case strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "rate limited"):
		return fmt.Errorf(
			"%w: Zoom's daily meeting-creation limit has been reached "+
				"for your account. You can either wait until 00:00 GMT "+
				"for the limit to reset, or paste a Zoom link manually "+
				"on the schedule and publish again",
			domain.ErrPublishRateLimited,
		)

	case strings.Contains(msg, "platform unavailable") ||
		strings.Contains(msg, "unavailable"):
		return fmt.Errorf(
			"%w: the video platform is temporarily unavailable. "+
				"Paste a Zoom link manually on the schedule, or try again "+
				"in a few minutes",
			domain.ErrPublishPlatformUnavailable,
		)

	case strings.Contains(msg, "not connected") ||
		strings.Contains(msg, "refresh token") ||
		strings.Contains(msg, "connection revoked") ||
		strings.Contains(msg, "needs reauthorization"):
		return fmt.Errorf(
			"%w: your Zoom connection is no longer valid. "+
				"Reconnect Zoom in Settings → Integrations, or paste a "+
				"link manually on the schedule",
			domain.ErrPublishConnectionInvalid,
		)

	default:
		return fmt.Errorf(
			"%w: the automatic meeting could not be created. "+
				"Paste a link manually on the schedule, or try again later. "+
				"(details: %v)",
			domain.ErrPublishMeetingCreationFailed,
			meetingErr,
		)
	}
}