// internal/modules/authorization/wire.go

package authorization

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(
	ProvideAuthorizationModule,
	ProvideAuthorizationService,
	ProvideEnforcer,
)

func ProvideAuthorizationModule(db *gorm.DB, cfg *config.Config) (*Module, error) {
	return NewModule(db, cfg)
}

func ProvideAuthorizationService(module *Module) *Service {
	return module.GetService()
}

func ProvideEnforcer(module *Module) *Enforcer {
	return module.GetEnforcer()
}