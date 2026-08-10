package eventhandler

import (
	"github.com/gofiber/fiber/v3"
)

func (h *EventHandler) RegisterRoutes(
	router fiber.Router,
	authMiddleware fiber.Handler,      // Authentication middleware (JWT)
	authzMiddleware fiber.Handler,     // Authorization middleware (Casbin)
) {
	// ============================================================
	// PUBLIC ROUTES (No auth required)
	// ============================================================
	public := router.Group("/events")
	{
		public.Get("/upcoming", h.GetUpcomingEvents)
		public.Get("/past", h.GetPastEvents)
		public.Get("/types", h.GetEventTypes)
		public.Get("/statuses", h.GetEventStatuses)
		public.Get("/search", h.SearchEvents)
		public.Get("/", h.ListEvents)
		public.Get("/:id", h.GetEvent)
		public.Get("/slug/:slug", h.GetEventBySlug)
		public.Get("/type/:type", h.GetEventsByType)
	}

	// ============================================================
	// PROTECTED ROUTES (Auth required)
	// ============================================================
	protected := router.Group("/accounts/:accountId/events")
	protected.Use(authMiddleware)        // ✅ Validate JWT first
	protected.Use(authzMiddleware)       // ✅ Then check permissions
	{
		protected.Get("/", h.GetEventsByAccount)
		protected.Post("/", h.CreateEvent)
		protected.Post("/with-image", h.CreateEventWithImage)
		protected.Post("/:eventId/image", h.UploadEventImage)
		protected.Post("/:eventId/certificate", h.UploadCertificateTemplate)
	}

	// ============================================================
	// EVENT MANAGEMENT (Auth required)
	// ============================================================
	eventRoutes := router.Group("/events")
	eventRoutes.Use(authMiddleware)        // ✅ Validate JWT first
	eventRoutes.Use(authzMiddleware)       // ✅ Then check permissions
	{
		eventRoutes.Put("/:id", h.UpdateEvent)
		eventRoutes.Delete("/:id", h.DeleteEvent)
		eventRoutes.Post("/:id/publish", h.PublishEvent)
		eventRoutes.Post("/:id/cancel", h.CancelEvent)
		eventRoutes.Post("/:id/complete", h.CompleteEvent)
	}
}