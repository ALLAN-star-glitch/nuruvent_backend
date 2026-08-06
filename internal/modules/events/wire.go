package events

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/media"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	ProvideEventsModule,
)

func ProvideEventsModule(
	cfg *config.Config,
	enforcer *authorization.Enforcer,
	permService *authorization.Service,
	mediaModule *media.Module,
) *Module {
	return NewModule(
		cfg,
		enforcer,
		permService,
		mediaModule.GetService(),
	)
}