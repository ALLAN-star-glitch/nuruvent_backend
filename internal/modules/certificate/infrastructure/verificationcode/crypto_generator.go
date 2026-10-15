// internal/modules/certificate/infrastructure/verificationcode/crypto_generator.go

package verificationcode

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// CodeLength is the number of random characters in a verification code.
//
// 16 characters over a 62-symbol alphabet yields ~95 bits of entropy —
// far above the minimum required by BRD NFR-SEC-001 and FR-CV-007, and
// effectively immune to enumeration.
const CodeLength = 16

// alphabet is the URL-safe character set used for verification codes.
// Matches what domain.VerificationCode.Validate accepts.
const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// CryptoGenerator produces cryptographically random verification codes
// using crypto/rand.
type CryptoGenerator struct{}

// NewCryptoGenerator constructs a CryptoGenerator.
func NewCryptoGenerator() *CryptoGenerator { return &CryptoGenerator{} }

// NewCode implements domain.VerificationCodeGenerator.
func (*CryptoGenerator) NewCode() (domain.VerificationCode, error) {
	b := make([]byte, CodeLength)
	max := big.NewInt(int64(len(alphabet)))

	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("verification code generation failed: %w", err)
		}
		b[i] = alphabet[n.Int64()]
	}

	return domain.NewVerificationCode(string(b))
}

// compile-time assertion
var _ domain.VerificationCodeGenerator = (*CryptoGenerator)(nil)