// internal/modules/video/service/zoom_signature.go

package service

import (
	"context"
	"fmt"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// GenerateMeetingSignature issues a signed JWT that authorizes the
// frontend to join the given meeting via the embedded Meeting SDK.
//
// The signature is bound to the meeting number and to a role: 0 for
// attendee, 1 for host. It expires in two hours.
//
// The user must have an active connection for the platform. This is
// not strictly required to sign a token, but it prevents issuing a
// signature to a user who cannot actually host or attend.
func (s *videoService) GenerateMeetingSignature(
	ctx context.Context,
	cmd GenerateMeetingSignatureCommand,
) (*videodomain.MeetingSignature, error) {
	if strings.TrimSpace(cmd.UserID) == "" {
		return nil, fmt.Errorf("%w: user id is required", videodomain.ErrUnauthorized)
	}
	if !cmd.Platform.IsValid() {
		return nil, fmt.Errorf("%w: invalid platform", videodomain.ErrUnsupportedPlatform)
	}
	if strings.TrimSpace(cmd.MeetingNumber) == "" {
		return nil, fmt.Errorf("%w: meeting number is required", videodomain.ErrInvalidMeeting)
	}
	if cmd.Role != 0 && cmd.Role != 1 {
		return nil, fmt.Errorf("%w: role must be 0 or 1", videodomain.ErrInvalidMeeting)
	}

	conn, err := s.deps.Connections.FindActiveByUserAndPlatform(ctx, cmd.UserID, cmd.Platform)
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, videodomain.ErrNotConnected
		}
		return nil, fmt.Errorf("generate signature: find connection: %w", err)
	}
	if conn.RevokedAt != nil {
		return nil, videodomain.ErrConnectionRevoked
	}

	client, err := s.deps.Clients.For(cmd.Platform)
	if err != nil {
		return nil, fmt.Errorf("generate signature: %w", err)
	}
	signer, ok := client.(videodomain.MeetingSDKSigner)
	if !ok {
		return nil, fmt.Errorf("%w: platform %s does not support the meeting SDK",
			videodomain.ErrCapabilityMissing, cmd.Platform)
	}

	signature, err := signer.SignedMeetingJWT(cmd.MeetingNumber, cmd.Role)
	if err != nil {
		return nil, fmt.Errorf("generate signature: %w", err)
	}

	return &videodomain.MeetingSignature{
		Signature: signature,
		SDKKey:    signer.ClientID(),
	}, nil
}