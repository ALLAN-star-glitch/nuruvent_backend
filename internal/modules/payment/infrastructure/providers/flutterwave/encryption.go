package flutterwave

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// encryptCardPayload encrypts card data using AES-256-GCM.
//
// Flutterwave requires this scheme:
//   - Key: the encryption key from the dashboard, used as-is (32 bytes)
//   - Nonce: exactly 12 bytes, random per encryption
//   - Output: base64( nonce || ciphertext || tag )
//
// The encrypted value is sent as the value of the card field in the
// charge request body.
func encryptCardPayload(plaintext, encryptionKey string) (string, error) {
	key := []byte(encryptionKey)
	if len(key) != 32 {
		return "", fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize()) // 12 bytes for GCM standard
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Concatenate nonce + ciphertext, then base64 encode.
	combined := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(combined), nil
}