// internal/modules/video/service/disconnect.go

package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// Disconnect revokes the user's active connection to a platform.
//
// Steps:
//  1. Load the user's active connection.
//  2. Ask the platform to revoke the tokens. Best-effort — a
//     failure here is logged but does not prevent the local
//     connection from being marked revoked.
//  3. Mark the local connection as revoked.
//
// Idempotent: if no active connection exists, returns nil.
func (s *videoService) Disconnect(
	ctx context.Context,
	cmd DisconnectCommand,
) error {
	if strings.TrimSpace(cmd.UserID) == "" {
		return fmt.Errorf("%w: user id is required", videodomain.ErrUnauthorized)
	}
	if !cmd.Platform.IsValid() {
		return fmt.Errorf("%w: invalid platform %q",
			videodomain.ErrUnsupportedPlatform, cmd.Platform)
	}

	conn, err := s.deps.Connections.FindActiveByUserAndPlatform(ctx, cmd.UserID, cmd.Platform)
	if err != nil {
		if errorsIsNotFound(err) {
			// Already disconnected. Idempotent success.
			return nil
		}
		return fmt.Errorf("find active connection: %w", err)
	}

	// Best-effort revocation at the platform. If the platform is
	// down or the token is already invalid, we still revoke locally.
	if oauth, err := s.deps.Clients.OAuthFor(cmd.Platform); err == nil {
		if err := oauth.RevokeAccess(ctx, conn); err != nil {
			log.Printf("[video] disconnect: platform revoke failed for user=%s platform=%s: %v",
				cmd.UserID, cmd.Platform, err)
			// fall through — revoke locally regardless
		}
	} else {
		log.Printf("[video] disconnect: no oauth client for platform=%s: %v",
			cmd.Platform, err)
	}

	now := s.deps.Clock.Now()

	err = s.deps.UnitOfWork.Do(ctx, func(repos videodomain.Repositories) error {
		// Re-fetch inside the transaction to get a fresh copy.
		// Another request may have revoked it while we were calling
		// the platform.
		current, err := repos.Connections.FindByID(ctx, conn.ID)
		if err != nil {
			if errorsIsNotFound(err) {
				return nil // already gone
			}
			return fmt.Errorf("reload connection: %w", err)
		}
		if current.RevokedAt != nil {
			return nil // already revoked
		}
		current.Revoke(now)
		if err := repos.Connections.Update(ctx, current); err != nil {
			return fmt.Errorf("mark revoked: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}