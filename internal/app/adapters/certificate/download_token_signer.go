// internal/app/adapters/certificate/download_token_signer.go

package certificate

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	certdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// DownloadTokenSignerAdapter implements certdomain.DownloadTokenSigner
// using HMAC-SHA256.
//
// Token format: base64url(payload) . base64url(hmac)
//
// Payload is a compact JSON object:
//
//	{"c":"cert-uuid","e":"user@example.com","x":1735689600}
type DownloadTokenSignerAdapter struct {
	secret []byte
}

// NewDownloadTokenSignerAdapter constructs the adapter.
func NewDownloadTokenSignerAdapter(secret string) *DownloadTokenSignerAdapter {
	return &DownloadTokenSignerAdapter{secret: []byte(secret)}
}

// downloadTokenPayload is the wire shape. Short keys keep the URL
// reasonably short.
type downloadTokenPayload struct {
	CertificateID string `json:"c"`
	Email         string `json:"e"`
	ExpiresAt     int64  `json:"x"`
}

// Sign implements certdomain.DownloadTokenSigner.
func (a *DownloadTokenSignerAdapter) Sign(
	_ context.Context,
	certificateID, email string,
	ttl time.Duration,
) (string, error) {
	if certificateID == "" || email == "" {
		return "", fmt.Errorf("certificate_id and email are required")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("ttl must be positive")
	}

	payload := downloadTokenPayload{
		CertificateID: certificateID,
		Email:         email,
		ExpiresAt:     time.Now().Add(ttl).Unix(),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	sig := a.sign(payloadB64)
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return payloadB64 + "." + sigB64, nil
}

// Verify implements certdomain.DownloadTokenSigner.
func (a *DownloadTokenSignerAdapter) Verify(
	_ context.Context,
	token string,
) (*certdomain.DownloadTokenClaims, error) {
	// Split.
	dot := -1
	for i := len(token) - 1; i >= 0; i-- {
		if token[i] == '.' {
			dot = i
			break
		}
	}
	if dot <= 0 || dot == len(token)-1 {
		return nil, certdomain.ErrInvalidDownloadToken
	}

	payloadB64 := token[:dot]
	sigB64 := token[dot+1:]

	// Verify signature.
	sig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return nil, certdomain.ErrInvalidDownloadToken
	}
	expected := a.sign(payloadB64)
	if !hmac.Equal(sig, expected) {
		return nil, certdomain.ErrInvalidDownloadToken
	}

	// Decode payload.
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, certdomain.ErrInvalidDownloadToken
	}

	var payload downloadTokenPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, certdomain.ErrInvalidDownloadToken
	}

	// Check expiry.
	expiresAt := time.Unix(payload.ExpiresAt, 0).UTC()
	if time.Now().UTC().After(expiresAt) {
		return nil, certdomain.ErrExpiredDownloadToken
	}

	return &certdomain.DownloadTokenClaims{
		CertificateID: payload.CertificateID,
		Email:         payload.Email,
		ExpiresAt:     expiresAt,
	}, nil
}

// sign computes HMAC-SHA256 over the payload.
func (a *DownloadTokenSignerAdapter) sign(payload string) []byte {
	h := hmac.New(sha256.New, a.secret)
	h.Write([]byte(payload))
	return h.Sum(nil)
}

// compile-time assertion
var _ certdomain.DownloadTokenSigner = (*DownloadTokenSignerAdapter)(nil)

// silence the unused import in case errors is only used above
var _ = errors.Is