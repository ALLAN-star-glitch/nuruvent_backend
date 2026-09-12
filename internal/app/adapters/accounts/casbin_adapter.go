// internal/app/adapters/accounts/casbin_adapter.go

package accounts

import (
	"context"
	"fmt"

	accountdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
)

// CasbinAdapter implements accountdomain.CasbinService using the auth
// module's enforcer.
type CasbinAdapter struct {
	enforcer *authorization.Enforcer
}

func NewCasbinAdapter(enforcer *authorization.Enforcer) accountdomain.CasbinService {
	return &CasbinAdapter{enforcer: enforcer}
}

func (a *CasbinAdapter) AssignAccountRole(ctx context.Context, accountID, userID, role string) error {
	domain := accountdomain.AccountDomain(accountID)
	if domain == "" {
		return fmt.Errorf("invalid account ID: %q", accountID)
	}
	_, err := a.enforcer.AddGroupingPolicy(userID, role, domain)
	return err
}

func (a *CasbinAdapter) RemoveAccountRole(ctx context.Context, accountID, userID string) error {
	domain := accountdomain.AccountDomain(accountID)
	if domain == "" {
		return fmt.Errorf("invalid account ID: %q", accountID)
	}
	// ptype=g, v0=userID, v1=*, v2=domain
	_, err := a.enforcer.RemoveFilteredGroupingPolicy(0, userID, "", domain)
	return err
}

func (a *CasbinAdapter) RemoveAllAccountRoles(ctx context.Context, accountID string) error {
	domain := accountdomain.AccountDomain(accountID)
	if domain == "" {
		return fmt.Errorf("invalid account ID: %q", accountID)
	}
	// ptype=g, v2=domain
	_, err := a.enforcer.RemoveFilteredGroupingPolicy(2, domain)
	return err
}