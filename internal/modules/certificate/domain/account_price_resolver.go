// internal/modules/certificate/domain/account_price_resolver.go

package domain

import "context"

// ============================================================
// OUTBOUND PORT: AccountPriceResolver
// ============================================================
//
// Resolves the per-account certificate price. The certificate module
// does not own the accounts table; this port is the seam.
//
// Implementations may return NuruventCertificatePriceMinor as a
// fallback when the account has not been configured.
type AccountPriceResolver interface {
	// CertificatePriceMinor returns the account's configured certificate
	// price in minor units. Zero means "free" — the caller must
	// distinguish "explicitly free" from "not configured". This port
	// returns the value as stored; the fallback to
	// NuruventCertificatePriceMinor is the caller's responsibility.
	CertificatePriceMinor(ctx context.Context, accountID string) (int64, error)
}