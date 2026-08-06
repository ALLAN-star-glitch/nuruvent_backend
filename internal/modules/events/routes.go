// internal/modules/events/routes.go

package events

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/events/eventhandler"
	"github.com/gofiber/fiber/v3"
)

func RegisterEventRoutes(
	router fiber.Router,
	handler *eventhandler.EventHandler,
	enforcer *authorization.Enforcer,
) {
	// ================================================
	// PUBLIC ROUTES (No auth required)
	// ================================================
	public := router.Group("/events")
	{
		public.Get("/upcoming", handler.GetUpcomingEvents)
		public.Get("/past", handler.GetPastEvents)
		public.Get("/types", handler.GetEventTypes)
		public.Get("/search", handler.SearchEvents)
		public.Get("/:id", handler.GetEvent)
		public.Get("/slug/:slug", handler.GetEventBySlug)
		public.Get("/type/:type", handler.GetEventsByType)
	}

	// ================================================
	// PROTECTED ROUTES (Auth required)
	// ================================================
	protected := router.Group("/businesses/:businessId/events")
	protected.Use(authorization.AuthorizationMiddleware(enforcer))
	{
		protected.Post("/", handler.CreateEvent)
		protected.Get("/", handler.GetBusinessEvents)
		protected.Post("/:eventId/image", handler.UploadEventImage)
		protected.Post("/:eventId/certificate", handler.UploadCertificateTemplate)
	}

	// ================================================
	// EVENT CRUD (Auth required)
	// ================================================
	eventRoutes := router.Group("/events")
	eventRoutes.Use(authorization.AuthorizationMiddleware(enforcer))
	{
		eventRoutes.Put("/:id", handler.UpdateEvent)
		eventRoutes.Delete("/:id", handler.DeleteEvent)
	}
}