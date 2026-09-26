
// internal/modules/video/infrastructure/crypto/cipher.go

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// AESGCMCipher encrypts tokens using AES-256-GCM.
//
// The key must be 32 bytes. Stored values are base64(nonce || ciphertext),
// safe to store as text.
//
// AES-GCM provides confidentiality and integrity — tampering with the
// ciphertext causes Decrypt to fail. That's the right primitive for
// short-lived secrets like OAuth tokens.
type AESGCMCipher struct {
	gcm cipher.AEAD
}

// NewAESGCMCipher constructs a cipher from a base64-encoded 32-byte key.
func NewAESGCMCipher(base64Key string) (*AESGCMCipher, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("decode encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	return &AESGCMCipher{gcm: gcm}, nil
}

// Encrypt returns base64(nonce || ciphertext) for the given plaintext.
// Empty input produces empty output.
func (c *AESGCMCipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}
	sealed := c.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt reverses Encrypt. Empty input produces empty output.
func (c *AESGCMCipher) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("%w: %v", videodomain.ErrInvalidCiphertext, err)
	}
	nonceSize := c.gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("%w: ciphertext too short", videodomain.ErrInvalidCiphertext)
	}
	nonce, sealed := raw[:nonceSize], raw[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", videodomain.ErrInvalidCiphertext, err)
	}
	return string(plaintext), nil
}

// Compile-time assertion.
var _ videodomain.TokenCipher = (*AESGCMCipher)(nil)