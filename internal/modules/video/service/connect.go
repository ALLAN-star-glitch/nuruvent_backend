// internal/modules/video/service/connect.go

package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// stateByteLen is the number of random bytes used for the OAuth state.
// 32 bytes → 43 chars in base64url, well above the 128-bit minimum.
const stateByteLen = 32

// BeginConnect starts the OAuth flow.
//
// Steps:
//  1. Validate inputs and platform.
//  2. Resolve the provider client and confirm it supports OAuth.
//  3. Generate a random, single-use state.
//  4. Persist the state, bound to the initiating user and platform.
//  5. Return the platform's authorize URL for the caller to redirect to.
//
// If the user already has an active connection for this platform,
// BeginConnect still succeeds — reconnecting overwrites the previous
// connection at the callback step. The frontend decides whether to
// prompt the user before calling this.
func (s *videoService) BeginConnect(
	ctx context.Context,
	cmd BeginConnectCommand,
) (*ConnectResult, error) {
	if strings.TrimSpace(cmd.UserID) == "" {
		return nil, fmt.Errorf("%w: user id is required", videodomain.ErrUnauthorized)
	}
	if !cmd.Platform.IsValid() {
		return nil, fmt.Errorf("%w: invalid platform %q",
			videodomain.ErrUnsupportedPlatform, cmd.Platform)
	}

	// Resolve the provider and confirm OAuth support.
	oauth, err := s.deps.Clients.OAuthFor(cmd.Platform)
	if err != nil {
		return nil, err
	}

	// Generate a single-use state.
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("generate state: %w", err)
	}

	now := s.deps.Clock.Now()
	record, err := videodomain.NewOAuthState(
		state,
		cmd.UserID,
		cmd.Platform,
		cmd.ReturnURL,
		now,
	)
	if err != nil {
		return nil, err
	}

	// Persist.
	if err := s.deps.OAuthStates.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("persist oauth state: %w", err)
	}

	return &ConnectResult{
		AuthorizeURL: oauth.OAuthAuthorizeURL(state),
		State:        state,
	}, nil
}

// generateState returns a cryptographically random URL-safe string.
func generateState() (string, error) {
	buf := make([]byte, stateByteLen) // This is a slice with length 32 bytes... buf is a slice of 32 bytes, which is the length we want for our random state.
	if _, err := rand.Read(buf); err != nil { // Fill the slice with random bytes. rand.Read returns the number of bytes read and an error. We ignore the number of bytes read because we know it will fill the entire slice, and we only care about the error.
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil // Encode the random bytes as a base64 URL-safe string. We use RawURLEncoding to avoid padding characters, which makes the string shorter and still URL-safe. Padding characters are characters like '=' that are used in standard base64 encoding to make the output length a multiple of 4. RawURLEncoding omits these characters, which is fine for our use case because we don't need the padding for decoding.
}
