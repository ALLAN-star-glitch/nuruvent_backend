// internal/modules/video/service/update_meeting.go

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// UpdateMeeting updates an existing meeting on the host's platform
// account.
//
// Steps:
//  1. Validate inputs.
//  2. Load the host's active connection for the requested platform.
//  3. Ensure the access token is fresh (refresh if near expiry).
//  4. Resolve the platform's MeetingProvisioner.
//  5. Ask the platform to update the meeting.
//  6. Load the local record, apply the new spec, persist.
//
// The JoinURL is not returned by Zoom's PATCH endpoint; the local
// record keeps its existing JoinURL unchanged. Attendees who already
// have the link are unaffected.
func (s *videoService) UpdateMeeting(
	ctx context.Context,
	cmd UpdateMeetingCommand,
) (*videodomain.Meeting, error) {
	if err := validateUpdateMeetingCommand(cmd); err != nil {
		return nil, err
	}

	// 1. Resolve the host's active connection.
	conn, err := s.deps.Connections.FindActiveByUserAndPlatform(
		ctx, cmd.UserID, cmd.Platform,
	)
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, videodomain.ErrNotConnected
		}
		return nil, fmt.Errorf("update meeting: find connection: %w", err)
	}
	if conn.RevokedAt != nil {
		return nil, videodomain.ErrConnectionRevoked
	}

	// 2. Ensure the access token is fresh.
	conn, err = s.ensureFreshToken(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("update meeting: %w", err)
	}

	// 3. Resolve the provisioner.
	provisioner, err := s.deps.Clients.ProvisionerFor(conn.Platform)
	if err != nil {
		return nil, fmt.Errorf("update meeting: %w", err)
	}

	// 4. Ask the platform. The client returns a partially-populated
	//    Meeting — ExternalID, JoinURL (empty on PATCH), Topic,
	//    StartTime, Duration, Timezone.
	platformMeeting, err := provisioner.UpdateMeeting(ctx, conn, cmd.ExternalID, cmd.Spec)
	if err != nil {
		return nil, fmt.Errorf("update meeting: platform call: %w", err)
	}
	if platformMeeting == nil {
		return nil, errors.New("update meeting: platform returned nil meeting")
	}

	// 5. Load the local record so we can preserve JoinURL, StartURL,
	//    Password, and CreatedAt. Zoom's PATCH response does not
	//    include these, so the existing row is the source of truth.
	local, err := s.deps.Meetings.FindByExternalID(ctx, conn.Platform, cmd.ExternalID)
	if err != nil {
		return nil, fmt.Errorf("update meeting: load local record: %w", err)
	}
	if local == nil {
		return nil, videodomain.ErrMeetingNotFound
	}

	// 6. Apply the new spec to the local record.
	local.Topic = cmd.Spec.Topic
	local.StartTime = cmd.Spec.StartTime
	local.Duration = cmd.Spec.Duration
	local.Timezone = cmd.Spec.Timezone
	local.UpdatedAt = s.deps.Clock.Now()
	// JoinURL, StartURL, Password, CreatedAt stay as they were.

	if err := s.deps.Meetings.Update(ctx, local); err != nil {
		return local, fmt.Errorf("update meeting: persist local record: %w", err)
	}

	return local, nil
}

// validateUpdateMeetingCommand checks required fields before any I/O.
func validateUpdateMeetingCommand(cmd UpdateMeetingCommand) error {
	if strings.TrimSpace(cmd.UserID) == "" {
		return fmt.Errorf("%w: user id is required", videodomain.ErrUnauthorized)
	}
	if !cmd.Platform.IsValid() {
		return fmt.Errorf("%w: invalid platform %q",
			videodomain.ErrUnsupportedPlatform, cmd.Platform)
	}
	if strings.TrimSpace(cmd.ExternalID) == "" {
		return fmt.Errorf("%w: external id is required", videodomain.ErrInvalidMeeting)
	}
	if cmd.Spec.HostUserID == "" {
		return fmt.Errorf("%w: spec.host_user_id is required",
			videodomain.ErrInvalidMeeting)
	}
	if cmd.Spec.HostUserID != cmd.UserID {
		return fmt.Errorf("%w: spec.host_user_id must match user_id",
			videodomain.ErrInvalidMeeting)
	}
	if err := cmd.Spec.Validate(); err != nil {
		return err
	}
	return nil
}