package events

import (
	"context"
	"errors"
	"strings"

	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	videoService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/service"
	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// VideoIdentityAdapter implements eventsdomain.VideoIdentityProvider
// by delegating to the video service's connection lookup.
type VideoIdentityAdapter struct {
	video videoService.Service
}

func NewVideoIdentityAdapter(
	video videoService.Service,
) eventsDomain.VideoIdentityProvider {
	return &VideoIdentityAdapter{video: video}
}

func (a *VideoIdentityAdapter) ExternalUserIDForPlatform(
	ctx context.Context,
	userID string,
	platform string,
) (string, error) {
	p := videodomain.Platform(platform)
	if !p.IsValid() {
		return "", nil
	}

	conn, err := a.video.GetConnection(ctx, userID, p)
	if err != nil {
		if errors.Is(err, videodomain.ErrConnectionNotFound) ||
			errors.Is(err, videodomain.ErrConnectionRevoked) {
			return "", nil
		}
		return "", err
	}
	if conn == nil || conn.ExternalUserID == "" {
		return "", nil
	}

	// Normalize Google user ids to the "users/<id>" form the
	// conference records API returns. The connection stores the bare
	// numeric id; the fetch's matcher does exact string equality, so
	// the stored value must match what Google sends.
	if p == videodomain.PlatformGoogleMeet &&
		!strings.HasPrefix(conn.ExternalUserID, "users/") {
		return "users/" + conn.ExternalUserID, nil
	}

	return conn.ExternalUserID, nil
}

// compile-time assertion
var _ eventsDomain.VideoIdentityProvider = (*VideoIdentityAdapter)(nil)