// internal/modules/video/service/zoom_zak.go

package service

import (
	"context"
	"fmt"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// FetchMeetingZAK returns a Zoom Access Key token for the given user's
// active connection.
//
// The ZAK authenticates the user as the meeting host inside the
// embedded Meeting SDK. It requires the user:read:zak scope on the
// OAuth app.
//
// The connection's access token is refreshed if near expiry before
// the Zoom call. If the refresh fails, the user is prompted to
// reconnect.
func (s *videoService) FetchMeetingZAK(
	ctx context.Context,
	cmd FetchMeetingZAKCommand,
) (*videodomain.MeetingZAK, error) {
	if strings.TrimSpace(cmd.UserID) == "" {
		return nil, fmt.Errorf("%w: user id is required", videodomain.ErrUnauthorized)
	}
	if !cmd.Platform.IsValid() {
		return nil, fmt.Errorf("%w: invalid platform", videodomain.ErrUnsupportedPlatform)
	}

	conn, err := s.deps.Connections.FindActiveByUserAndPlatform(ctx, cmd.UserID, cmd.Platform)
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, videodomain.ErrNotConnected
		}
		return nil, fmt.Errorf("fetch zak: find connection: %w", err)
	}
	if conn.RevokedAt != nil {
		return nil, videodomain.ErrConnectionRevoked
	}

	conn, err = s.ensureFreshToken(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("fetch zak: %w", err)
	}

	client, err := s.deps.Clients.For(cmd.Platform)
	if err != nil {
		return nil, fmt.Errorf("fetch zak: %w", err)
	}
	signer, ok := client.(videodomain.MeetingSDKSigner)
	if !ok {
		return nil, fmt.Errorf("%w: platform %s does not support the meeting SDK",
			videodomain.ErrCapabilityMissing, cmd.Platform)
	}

	zak, err := signer.FetchZAK(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("fetch zak: %w", err)
	}

	return &videodomain.MeetingZAK{ZAK: zak}, nil
}