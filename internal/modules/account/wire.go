// internal/modules/account/wire.go

package account

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/account/accountrepo"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/account/accountservice"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet defines all dependencies for the account module
var ProviderSet = wire.NewSet(
	ProvideAccountRepository,
	ProvideAccountService,
)

// ProvideAccountRepository provides the account repository
func ProvideAccountRepository(db *gorm.DB) *accountrepo.Repository {
	return accountrepo.NewRepository(db)
}

// ProvideAccountService provides the account service
func ProvideAccountService(repo *accountrepo.Repository) *accountservice.Service {
	return accountservice.NewService(repo)
}