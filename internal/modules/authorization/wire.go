// internal/modules/authorization/wire.go

package authorization

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet defines all dependencies for the authorization module
var ProviderSet = wire.NewSet(
	ProvideAuthorizationModule,
)

// ProvideAuthorizationModule provides the authorization module
func ProvideAuthorizationModule(db *gorm.DB, cfg *config.Config) (*Module, error) {
	return NewModule(db, cfg)
}