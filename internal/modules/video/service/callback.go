// internal/modules/video/service/callback.go

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// HandleCallback is called when the platform redirects back to
// Nuruvent after the user authorizes (or declines).
//
// Steps:
//  1. If the platform returned an error, reject.
//  2. Consume the state atomically — rejects unknown, consumed, and
//     expired states.
//  3. Resolve the provider client for the state's platform.
//  4. Exchange the authorization code for tokens + identity.
//  5. In a single transaction:
//     a. Revoke any existing active connection for the same user
//        and platform.
//     b. Insert the new connection.
//  6. Return the connection and the state's return URL.
func (s *videoService) HandleCallback(
	ctx context.Context,
	cmd CallbackCommand,
) (*ConnectionResult, error) {
	if strings.TrimSpace(cmd.State) == "" {
		return nil, fmt.Errorf("%w: state is required", videodomain.ErrInvalidOAuthState)
	}

	// The platform may return an error instead of a code — e.g. the
	// user clicked "Deny". Zoom sends error=access_denied.
	if cmd.Error != "" {
		return nil, fmt.Errorf("%w: %s", videodomain.ErrUnauthorized, cmd.Error)
	}
	if strings.TrimSpace(cmd.Code) == "" {
		return nil, fmt.Errorf("%w: code is required", videodomain.ErrInvalidOAuthState)
	}

	now := s.deps.Clock.Now()

	// 1. Consume the state. Rejects unknown, consumed, and expired.
	state, err := s.deps.OAuthStates.Consume(ctx, cmd.State, now)
	if err != nil {
		return nil, err
	}

	// 2. Resolve the provider client.
	oauth, err := s.deps.Clients.OAuthFor(state.Platform)
	if err != nil {
		return nil, err
	}

	// 3. Exchange the code. The client does both the token exchange
	//    and the user info lookup.
	externalUser, tokens, err := oauth.ExchangeCode(ctx, cmd.Code)
	if err != nil {
		return nil, err
	}

	// 4. Build the connection. The service generates the ID and
	//    timestamps; the client provided the identity and tokens.
	conn, err := videodomain.NewConnection(
		s.deps.IDs.NewID(),
		state.UserID,
		state.Platform,
		externalUser.ID,
		externalUser.Email,
		externalUser.OrgID,
		tokens.AccessToken,
		tokens.RefreshToken,
		tokens.ExpiresAt,
		tokens.Scopes,
		now,
	)
	if err != nil {
		return nil, err
	}

	// 5. Persist. Any existing active connection for the same
	//    (user, platform) is revoked first — the partial unique
	//    index would otherwise reject the insert.
	err = s.deps.UnitOfWork.Do(ctx, func(repos videodomain.Repositories) error {
		// Revoke existing, if any. Missing is fine.
		existing, err := repos.Connections.FindActiveByUserAndPlatform(
			ctx, state.UserID, state.Platform,
		)
		switch {
		case err == nil && existing != nil:
			existing.Revoke(now)
			if err := repos.Connections.Update(ctx, existing); err != nil {
				return fmt.Errorf("revoke previous connection: %w", err)
			}
		case err != nil && !errors.Is(err, videodomain.ErrConnectionNotFound):
			return fmt.Errorf("find existing connection: %w", err)
		}

		if err := repos.Connections.Create(ctx, conn); err != nil {
			return fmt.Errorf("create connection: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ConnectionResult{
		Connection: conn,
		ReturnURL:  state.ReturnURL,
	}, nil
}