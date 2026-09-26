// internal/modules/video/videodomain/crypto.go

package videodomain

// TokenCipher encrypts and decrypts OAuth tokens at rest.
//
// Implemented in the infrastructure layer. The repository depends on
// this interface, never on a concrete algorithm.
//
// Implementations must produce ciphertext that can be stored as text
// (base64, hex, etc.) and must be safe for concurrent use.
type TokenCipher interface {
	// Encrypt returns a reversible ciphertext for the given plaintext.
	// Empty input produces empty output.
	Encrypt(plaintext string) (string, error)

	// Decrypt returns the plaintext for a ciphertext produced by
	// Encrypt. Empty input produces empty output.
	//
	// Returns ErrInvalidCiphertext if the input is malformed or the
	// key doesn't match.
	Decrypt(ciphertext string) (string, error)
}