// internal/modules/certificate/domain/download_token_signer.go

package domain

import (
	"context"
	"time"
)

// ============================================================
// OUTBOUND PORT: DownloadTokenSigner
// ============================================================
//
// Signs and verifies the download tokens embedded in certificate
// emails. The token authorizes a single certificate download without
// requiring the recipient to authenticate — necessary for guest
// recipients who have no Nuruvent account.
//
// Tokens are time-limited. When they expire, the recipient must
// either sign in to their account (if they have one) or request a
// fresh link from the organizer.
type DownloadTokenSigner interface {
	// Sign produces a token that authorizes a download of the given
	// certificate by the given email address, valid for ttl.
	Sign(
		ctx context.Context,
		certificateID, email string,
		ttl time.Duration,
	) (string, error)

	// Verify parses and validates a token. Returns ErrInvalidDownloadToken
	// or ErrExpiredDownloadToken on failure.
	Verify(
		ctx context.Context,
		token string,
	) (*DownloadTokenClaims, error)
}

// DownloadTokenClaims is what a verified token contains.
type DownloadTokenClaims struct {
	CertificateID string
	Email         string
	ExpiresAt     time.Time
}