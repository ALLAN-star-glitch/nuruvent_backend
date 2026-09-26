
// internal/modules/attendance/infrastructure/token/generator.go

package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// SHA256Generator produces opaque random tokens and hashes them with
// SHA-256.
//
// The raw token is 32 bytes of cryptographic randomness, base64url
// encoded with a "nrt_" prefix. Only the SHA-256 hash is stored; the
// raw token is returned to the caller once and never persisted.
//
// SHA-256 (rather than bcrypt) is correct here because tokens have
// 256 bits of entropy — brute force is impossible regardless of hash
// speed, and we need the hash to be deterministic for lookups.
type SHA256Generator struct{}

// NewSHA256Generator constructs the default token generator.
func NewSHA256Generator() *SHA256Generator {
	return &SHA256Generator{}
}

// Generate returns (rawToken, tokenHash).
//
// The two values are computed together so the caller cannot
// accidentally store the raw token instead of its hash.
func (g *SHA256Generator) Generate() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("read random bytes: %w", err)
	}
	raw := "nrt_" + base64.RawURLEncoding.EncodeToString(buf)
	return raw, g.Hash(raw), nil
}

// Hash returns the hex-encoded SHA-256 of a raw token.
func (g *SHA256Generator) Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}