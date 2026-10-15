// internal/app/adapters/certificate/account_price_resolver.go

package certificate

import (
	"context"

	certdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	"gorm.io/gorm"
)

// AccountPriceResolverAdapter reads certificate_price_minor from the
// accounts table. It queries the DB directly — the account module
// does not expose a service method for this yet.
type AccountPriceResolverAdapter struct {
	db *gorm.DB
}

// NewAccountPriceResolverAdapter constructs an AccountPriceResolverAdapter.
func NewAccountPriceResolverAdapter(db *gorm.DB) *AccountPriceResolverAdapter {
	return &AccountPriceResolverAdapter{db: db}
}

// CertificatePriceMinor implements certdomain.AccountPriceResolver.
func (a *AccountPriceResolverAdapter) CertificatePriceMinor(
	ctx context.Context,
	accountID string,
) (int64, error) {
	if accountID == "" {
		return certdomain.NuruventCertificatePriceMinor, nil
	}

	var result struct {
		PriceMinor int64
	}

	err := a.db.WithContext(ctx).
		Table("accounts").
		Select("certificate_price_minor AS price_minor").
		Where("id = ?", accountID).
		Scan(&result).Error
	if err != nil {
		return 0, err
	}
	if result.PriceMinor < 0 {
		// Defensive: a negative price is a data error; use the default.
		return certdomain.NuruventCertificatePriceMinor, nil
	}
	return result.PriceMinor, nil
}

// compile-time assertion
var _ certdomain.AccountPriceResolver = (*AccountPriceResolverAdapter)(nil)