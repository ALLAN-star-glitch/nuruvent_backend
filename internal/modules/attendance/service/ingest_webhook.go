// internal/modules/attendance/service/ingest_webhook.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// IngestWebhook handles an inbound webhook from a video provider by
// parsing the raw payload and then processing the resulting event.
//
// Kept for callers that have only the raw bytes (e.g. tests, workers).
// Production HTTP handlers should call ProcessWebhookEvent directly,
// after parsing in the delivery layer so provider-specific verification
// handshakes (like Zoom's endpoint.url_validation) can be intercepted.
func (s *attendanceService) IngestWebhook(
	ctx context.Context,
	provider attendance.SessionProvider,
	payload []byte,
	headers map[string]string,
) error {
	adapter, ok := s.deps.Providers[provider]
	if !ok {
		return fmt.Errorf("%w: no adapter for provider %q", attendance.ErrInvalidSession, provider)
	}

	event, err := adapter.ParseWebhook(ctx, payload, headers)
	if err != nil {
		return fmt.Errorf("parse webhook: %w", err)
	}

	return s.ProcessWebhookEvent(ctx, provider, event)
}

// ProcessWebhookEvent handles an already-parsed, already-verified
// webhook event.
//
// Flow:
//  1. Validate the event shape.
//  2. Session-level events (meeting.started, meeting.ended) branch off
//     early — they have no participant to match.
//  3. Resolve the session by (provider, meeting_id).
//  4. Match the participant to an attendee registered for that session.
//  5. Record the join or leave.
//  6. Recompute session statuses once.
//
// Unmatched events are logged and ignored. Webhooks are noisy;
// returning an error would trigger provider redelivery, and
// redelivery can't fix an unmatched event.
func (s *attendanceService) ProcessWebhookEvent(
	ctx context.Context,
	provider attendance.SessionProvider,
	event *attendance.WebhookEvent,
) error {
	if event == nil {
		return fmt.Errorf("%w: nil event", attendance.ErrInvalidSession)
	}
	if event.ProviderMeetingID == "" {
		return fmt.Errorf("%w: missing meeting id", attendance.ErrInvalidSession)
	}
	if !event.EventType.IsValid() {
		return fmt.Errorf("%w: unknown event type %q", attendance.ErrInvalidSession, event.EventType)
	}

	// Session-level events have no participant. Handle them before
	// the per-participant path below.
	switch event.EventType {
	case attendance.WebhookEventStarted:
		return s.handleMeetingStarted(ctx, provider, event)
	case attendance.WebhookEventEnded:
		return s.handleMeetingEnded(ctx, provider, event)
	}

	// Resolve the session.
	session, err := s.findSessionByProviderMeeting(ctx, provider, event.ProviderMeetingID)
	if err != nil {
		if errors.Is(err, attendance.ErrSessionNotFound) {
			log.Printf("[attendance] ingest webhook: no session for provider=%s meeting=%s",
				provider, event.ProviderMeetingID)
			return nil
		}
		return err
	}

	// Match the participant to an attendee.
	attendee, err := s.matchAttendee(ctx, session.ID, event.ParticipantEmail)
	if err != nil {
		return err
	}
	if attendee == nil {
		log.Printf("[attendance] ingest webhook: unmatched participant provider=%s session=%s email=%q name=%q",
			provider, session.ID, event.ParticipantEmail, event.ParticipantName)
		return nil
	}

	// Record the join or leave.
	switch event.EventType {
	case attendance.WebhookEventJoined:
		_, err = s.RecordJoin(ctx, RecordJoinCommand{
			AttendeeID: attendee.ID,
			SessionID:  session.ID,
			JoinTime:   event.OccurredAt,
			Source:     sourceForProvider(provider),
		})
		if err != nil {
			return fmt.Errorf("record join: %w", err)
		}
	case attendance.WebhookEventLeft:
		err = s.RecordLeave(ctx, RecordLeaveCommand{
			AttendeeID: attendee.ID,
			SessionID:  session.ID,
			LeaveTime:  event.OccurredAt,
			Source:     sourceForProvider(provider),
		})
		if err != nil {
			return fmt.Errorf("record leave: %w", err)
		}
	default:
		return fmt.Errorf("%w: unhandled event type %q", attendance.ErrInvalidSession, event.EventType)
	}

	// Recompute once per delivery.
	if err := s.RecomputeSessionStatuses(ctx, session.ID); err != nil {
		log.Printf("[attendance] ingest webhook: recompute failed session=%s err=%v", session.ID, err)
	}

	return nil
}

// handleMeetingStarted processes a session-level meeting.started event.
//
// There is no participant to match and no row-level attendance to
// write — this fires once when the host starts the meeting, before
// any participant joins. What this hook is for:
//
//   - Flipping the session status to "live" so the UI can show a live
//     indicator.
//   - Future: emit real-time notifications, send "meeting is live"
//     messages to registered attendees, etc.
//
// Best-effort: a session we've never seen before logs and returns nil,
// because that's not a delivery error — the webhook is simply for a
// meeting the attendance module doesn't track.
func (s *attendanceService) handleMeetingStarted(
	ctx context.Context,
	provider attendance.SessionProvider,
	event *attendance.WebhookEvent,
) error {
	session, err := s.findSessionByProviderMeeting(ctx, provider, event.ProviderMeetingID)
	if err != nil {
		if errors.Is(err, attendance.ErrSessionNotFound) {
			log.Printf("[attendance] meeting.started: no session for provider=%s meeting=%s",
				provider, event.ProviderMeetingID)
			return nil
		}
		return err
	}

	if err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		return repos.Sessions.UpdateStatus(ctx, session.ID, attendance.SessionStatusLive)
	}); err != nil {
		log.Printf("[attendance] meeting.started: status update failed session=%s err=%v",
			session.ID, err)
	}

	log.Printf("[attendance] meeting.started: session=%s is live", session.ID)
	return nil
}

// handleMeetingEnded processes a session-level meeting.ended event.
//
// There is no participant to match and no row-level attendance to
// write — join/leave events already captured that. What this hook is
// for:
//
//   - Marking the session as finished (so the UI can stop showing it
//     as "in progress").
//   - Running a final recompute so any stale rollups get refreshed.
//   - Any future "close out the meeting" side effects.
//
// Best-effort: a session we've never seen before logs and returns nil,
// because that's not a delivery error — the webhook is simply for a
// meeting the attendance module doesn't track.
func (s *attendanceService) handleMeetingEnded(
	ctx context.Context,
	provider attendance.SessionProvider,
	event *attendance.WebhookEvent,
) error {
	session, err := s.findSessionByProviderMeeting(ctx, provider, event.ProviderMeetingID)
	if err != nil {
		if errors.Is(err, attendance.ErrSessionNotFound) {
			log.Printf("[attendance] meeting.ended: no session for provider=%s meeting=%s",
				provider, event.ProviderMeetingID)
			return nil
		}
		return err
	}

	// Final recompute so per-session rollups reflect any last
	// join/leave pairs that arrived just before the end event.
	if err := s.RecomputeSessionStatuses(ctx, session.ID); err != nil {
		log.Printf("[attendance] meeting.ended: recompute failed session=%s err=%v",
			session.ID, err)
	}

	if err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		return repos.Sessions.UpdateStatus(ctx, session.ID, attendance.SessionStatusEnded)
	}); err != nil {
		log.Printf("[attendance] meeting.ended: status update failed session=%s err=%v",
			session.ID, err)
	}

	log.Printf("[attendance] meeting.ended: finalized session=%s", session.ID)
	return nil
}




// findSessionByProviderMeeting loads a session by provider and meeting
// ID.
func (s *attendanceService) findSessionByProviderMeeting(
	ctx context.Context,
	provider attendance.SessionProvider,
	meetingID string,
) (*attendance.Session, error) {
	var session *attendance.Session

	err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		var txErr error
		session, txErr = repos.Sessions.FindByProviderMeetingID(ctx, provider, meetingID)
		return txErr
	})
	if err != nil {
		return nil, err
	}
	return session, nil
}

// matchAttendee finds the attendee registered for a session whose
// email matches the given address.
//
// Returns (nil, nil) if no attendee matches — that's a soft failure
// (unmatched participant), not an error.
func (s *attendanceService) matchAttendee(
	ctx context.Context,
	sessionID string,
	email string,
) (*attendance.Attendee, error) {
	if email == "" {
		return nil, nil
	}

	var matched *attendance.Attendee

	err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		candidates, err := repos.Attendees.FindByEmail(ctx, email)
		if err != nil {
			return fmt.Errorf("find attendees by email: %w", err)
		}
		if len(candidates) == 0 {
			return nil
		}

		var sessionAttendees []*attendance.Attendee
		for _, c := range candidates {
			_, err := repos.SessionStatuses.FindByAttendeeSession(ctx, c.ID, sessionID)
			if err == nil {
				sessionAttendees = append(sessionAttendees, c)
				continue
			}
			if errors.Is(err, attendance.ErrStatusNotFound) {
				continue
			}
			return fmt.Errorf("check session registration: %w", err)
		}

		switch len(sessionAttendees) {
		case 0:
			return nil
		case 1:
			matched = sessionAttendees[0]
			return nil
		default:
			log.Printf("[attendance] ingest webhook: ambiguous match session=%s email=%q candidates=%d",
				sessionID, email, len(sessionAttendees))
			return nil
		}
	})
	if err != nil {
		return nil, err
	}

	return matched, nil
}

// sourceForProvider maps a session provider to the source constant
// used on attendance records.
func sourceForProvider(p attendance.SessionProvider) attendance.AttendanceSource {
	switch p {
	case attendance.ProviderZoom:
		return attendance.SourceZoomWebhook
	case attendance.ProviderGoogleMeet:
		return attendance.SourceGoogleEvent
	default:
		return attendance.SourceHostManual
	}
}