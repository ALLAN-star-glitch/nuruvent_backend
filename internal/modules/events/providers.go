// internal/modules/events/providers.go

package events

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/delivery/eventhandler"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
)

var ProviderSet = wire.NewSet(

	// Repository
	postgres.NewPostgresRepository,

	// Service
	service.NewService,

	// Handler
	eventhandler.NewEventHandler,
)