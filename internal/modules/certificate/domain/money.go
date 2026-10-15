// internal/modules/certificate/domain/money.go

package domain

// NuruventCertificatePriceMinor is the fallback price for a certificate
// when the account does not override it. 6500 = KES 65.00.
//
// The account's certificate_price_minor column (added in migration 145)
// supersedes this constant. The constant exists so the service can
// price a certificate even if the account lookup fails.
const NuruventCertificatePriceMinor int64 = 6500

// DefaultCurrency is the currency certificates are priced in.
const DefaultCurrency = "KES"

// Money is a value object for monetary amounts. All amounts are in
// minor units (cents) to avoid floating-point drift.
type Money struct {
	AmountMinor int64
	Currency    string
}

// IsFree reports whether the amount is zero.
func (m Money) IsFree() bool { return m.AmountMinor == 0 }

// NewMoney constructs a Money, defaulting the currency to KES.
func NewMoney(amountMinor int64, currency string) Money {
	if currency == "" {
		currency = DefaultCurrency
	}
	return Money{AmountMinor: amountMinor, Currency: currency}
}