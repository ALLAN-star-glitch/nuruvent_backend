// internal/modules/events/providers.go

package events

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/delivery/eventhandler"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/infrastructure" // ← ADD
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/infrastructure/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
)

var ProviderSet = wire.NewSet(

	// Repository
	postgres.NewPostgresRepository,

	// AI adapter — implements service.AIService using shared/ai.Client
	infrastructure.NewEventAIAdapter,  // ← ADD

	// Service
	service.NewService,

	// Handler
	eventhandler.NewEventHandler,
)