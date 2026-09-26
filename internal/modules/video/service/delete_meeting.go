// internal/modules/video/service/delete_meeting.go

package service

import (
	"context"
	"fmt"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// DeleteMeeting removes a meeting from the platform and from local
// storage.
//
// Steps:
//  1. Validate inputs.
//  2. Load the meeting and verify ownership.
//  3. Resolve the host's current connection for the meeting's
//     platform. If the connection is gone or revoked, skip the
//     platform call and delete locally — nothing to authenticate with.
//  4. Ask the platform to delete the meeting. Best-effort: a failure
//     here is logged at the delivery layer and does not prevent local
//     deletion, because a meeting the host has already abandoned
//     shouldn't block their ability to remove the event from Nuruvent.
//  5. Delete the local record.
//
// Idempotent: calling with an ID that no longer exists returns nil.
func (s *videoService) DeleteMeeting(
	ctx context.Context,
	cmd DeleteMeetingCommand,
) error {
	if strings.TrimSpace(cmd.MeetingID) == "" {
		return fmt.Errorf("%w: meeting id is required", videodomain.ErrInvalidMeeting)
	}
	if strings.TrimSpace(cmd.UserID) == "" {
		return fmt.Errorf("%w: user id is required", videodomain.ErrUnauthorized)
	}

	// 1. Load the meeting.
	meeting, err := s.deps.Meetings.FindByID(ctx, cmd.MeetingID)
	if err != nil {
		if errorsIsMeetingNotFound(err) {
			return nil // idempotent
		}
		return fmt.Errorf("delete meeting: load: %w", err)
	}

	// 2. Ownership check — FR-V-047/048/049/308.
	if meeting.UserID != cmd.UserID {
		return fmt.Errorf("%w: meeting does not belong to user",
			videodomain.ErrForbidden)
	}

	// 3. Resolve the host's current connection for this platform.
	//    Meetings don't carry a connection FK — they belong to users.
	//    If the host has cycled through several Zoom connections, we
	//    use whatever is currently active, since only live credentials
	//    can reach the platform.
	conn, err := s.deps.Connections.FindActiveByUserAndPlatform(
		ctx, cmd.UserID, meeting.Platform,
	)
	if err != nil {
		if errorsIsNotFound(err) {
			// No active connection. Nothing to call the platform with.
			return s.deleteLocalMeeting(ctx, meeting.ID)
		}
		return fmt.Errorf("delete meeting: find connection: %w", err)
	}
	if conn.RevokedAt != nil {
		return s.deleteLocalMeeting(ctx, meeting.ID)
	}

	// 4. Best-effort platform delete.
	provisioner, err := s.deps.Clients.ProvisionerFor(meeting.Platform)
	if err != nil {
		// Platform not provisionable (shouldn't happen — we created
		// the meeting on this platform — but be graceful). Delete
		// locally and return.
		return s.deleteLocalMeeting(ctx, meeting.ID)
	}

	// Refresh before use. A refresh failure means the platform call
	// will fail too, so fall through to local delete.
	fresh, ferr := s.ensureFreshToken(ctx, conn)
	if ferr == nil {
		conn = fresh
		if derr := provisioner.DeleteMeeting(ctx, conn, meeting.ExternalID); derr != nil {
			// Delivery layer logs this. We deliberately swallow it —
			// the local record must go regardless of platform
			// outcome, because the alternative is a stuck meeting
			// the host can't remove.
			_ = derr
		}
	}

	// 5. Remove locally.
	return s.deleteLocalMeeting(ctx, meeting.ID)
}

// deleteLocalMeeting removes the meeting row. Runs inside a
// UnitOfWork so the operation is atomic with any future side effects.
func (s *videoService) deleteLocalMeeting(ctx context.Context, meetingID string) error {
	return s.deps.UnitOfWork.Do(ctx, func(repos videodomain.Repositories) error {
		if err := repos.Meetings.Delete(ctx, meetingID); err != nil {
			if errorsIsMeetingNotFound(err) {
				return nil // already gone
			}
			return fmt.Errorf("delete meeting: local delete: %w", err)
		}
		return nil
	})
}