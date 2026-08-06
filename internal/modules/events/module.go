// internal/modules/events/module.go

package events

import (
	"context"
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/database"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/events/eventhandler"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/events/eventrepo"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/events/eventservice"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/media/mediaservice"
	"github.com/gofiber/fiber/v3"
)

type Module struct {
	eventHandler *eventhandler.EventHandler
	eventService *eventservice.EventService
}

func NewModule(
	cfg *config.Config,
	enforcer *authorization.Enforcer,
	permService *authorization.Service,
	mediaService *mediaservice.MediaService,
) *Module {
	db := database.GetDB()

	// Initialize repository
	eventRepo := eventrepo.NewEventRepository(db)

	// Initialize service with media service injected
	eventService := eventservice.NewEventService(
		eventRepo,
		enforcer,
		permService,
		mediaService, // Inject media service
	)

	// Initialize handler
	eventHandler := eventhandler.NewEventHandler(eventService)

	return &Module{
		eventHandler: eventHandler,
		eventService: eventService,
		
	}
}

// SetupRoutes registers all event routes
func (m *Module) SetupRoutes(
	router fiber.Router,
	enforcer *authorization.Enforcer,
) {
	RegisterEventRoutes(
		router,
		m.eventHandler,
		enforcer,
	)
}

// GetEventService returns the event service
func (m *Module) GetEventService() *eventservice.EventService {
	return m.eventService
}

// GetEventHandler returns the event handler
func (m *Module) GetEventHandler() *eventhandler.EventHandler {
	return m.eventHandler
}

func (m *Module) Init(ctx context.Context) error {
	log.Println("Events module initialized")
	return nil
}

func (m *Module) Close() {
	log.Println("Events module closed")
}