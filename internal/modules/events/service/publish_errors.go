package service

import (
	"fmt"
	"strings"
)

// translatePublishError converts a validation error into a user-facing
// message that explains *why* the publish failed and what the user can
// do about it.
//
// validationErr is the error from event.ValidateForPublish().
// meetingErr is the error from attachVideoMeetings(), if any.
//
// When meetingErr is nil, the validation error is returned unchanged —
// the user's event just doesn't satisfy the rules, and the existing
// message is appropriate.
//
// When meetingErr is non-nil, we inspect its message to determine the
// cause and rewrite the error with an actionable instruction. This
// keeps the events module free of imports from the video module's
// domain package; the coupling is by message convention only.
func (s *eventService) translatePublishError(validationErr, meetingErr error) error {
	if meetingErr == nil {
		return fmt.Errorf("cannot publish event: %w", validationErr)
	}

	msg := meetingErr.Error()

	switch {
	case strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "rate limited"):
		return fmt.Errorf(
			"cannot publish: Zoom's daily meeting-creation limit has been " +
				"reached for your account. You can either wait until 00:00 GMT " +
				"for the limit to reset, or paste a Zoom link manually on the " +
				"schedule and publish again.",
		)

	case strings.Contains(msg, "platform unavailable") ||
		strings.Contains(msg, "unavailable"):
		return fmt.Errorf(
			"cannot publish: the video platform is temporarily unavailable. " +
				"Paste a Zoom link manually on the schedule, or try again " +
				"in a few minutes.",
		)

	case strings.Contains(msg, "not connected") ||
		strings.Contains(msg, "refresh token") ||
		strings.Contains(msg, "connection revoked") ||
		strings.Contains(msg, "needs reauthorization"):
		return fmt.Errorf(
			"cannot publish: your Zoom connection is no longer valid. " +
				"Reconnect Zoom in Settings → Integrations, or paste a link " +
				"manually on the schedule.",
		)

	default:
		return fmt.Errorf(
			"cannot publish: the automatic meeting could not be created. " +
				"Paste a link manually on the schedule, or try again later. " +
				"(details: %v)", meetingErr,
		)
	}
}