

package events

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/delivery/eventhandler"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
)

var ProviderSet = wire.NewSet(
	postgres.NewPostgresRepository,
	service.NewService,
	eventhandler.NewEventHandler,
)