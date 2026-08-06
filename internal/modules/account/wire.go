// internal/modules/account/wire.go

package account

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/account/accountrepo"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/account/accountservice"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(
	ProvideAccountRepository,
	ProvideAccountService,
	ProvideAccountModule,
)

func ProvideAccountRepository(db *gorm.DB) *accountrepo.Repository {
	return accountrepo.NewRepository(db)
}

func ProvideAccountService(repo *accountrepo.Repository) accountservice.Service {
	return accountservice.NewService(repo)
}

func ProvideAccountModule(svc accountservice.Service) *Module {
	return &Module{
		service: svc,
	}
}