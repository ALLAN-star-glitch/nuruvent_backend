// internal/app/adapters/certificate/account_reader_adapter.go

package certificate

import (
	"context"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
)

type AccountReaderAdapter struct {
	svc accountService.Service
}

func NewAccountReaderAdapter(svc accountService.Service) *AccountReaderAdapter {
	return &AccountReaderAdapter{svc: svc}
}

func (a *AccountReaderAdapter) GetAccountName(ctx context.Context, accountID string) (string, error) {
	account, err := a.svc.GetAccountByID(ctx, accountID)
	if err != nil {
		return "", fmt.Errorf("failed to get account %s: %w", accountID, err)
	}
	if account == nil {
		return "", fmt.Errorf("account %s not found", accountID)
	}

	// Prefer display_name, fall back to name.
	if account.DisplayName != "" {
		return account.DisplayName, nil
	}
	return account.Name, nil
}

var _ domain.AccountReader = (*AccountReaderAdapter)(nil)