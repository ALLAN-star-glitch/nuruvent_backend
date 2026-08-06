package server

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/events"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/media"
	"github.com/google/wire"
	"github.com/gofiber/fiber/v3"
)

var ProviderSet = wire.NewSet(
	ProvideSetupRoutes,
)

func ProvideSetupRoutes(
	app *fiber.App,
	cfg *config.Config,
	authHandler *auth.Handler,
	enforcer *authorization.Enforcer,
	mediaModule *media.Module,
	eventsModule *events.Module,
) error {
	SetupRoutes(
		app,
		cfg,
		authHandler,
		enforcer,
		mediaModule,
		eventsModule,
	)
	return nil
}