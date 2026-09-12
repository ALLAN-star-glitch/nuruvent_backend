package eventhandler

import (
	"github.com/gofiber/fiber/v3"
)

func (h *EventHandler) RegisterRoutes(
	router fiber.Router,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
) {
	// ============================================================
	// 1. PUBLIC ROUTES
	// ============================================================
	public := router.Group("/events")
	{
		public.Get("/upcoming", h.GetUpcomingEvents)
		public.Get("/past", h.GetPastEvents)
		public.Get("/search", h.SearchEvents)   // ⬅️ public search (no token assumed)
		public.Get("/types", h.GetEventTypes)
		public.Get("/statuses", h.GetEventStatuses)
		public.Get("/categories", h.GetCategories)
		public.Get("/ticket-types", h.GetTicketTypes)
		public.Get("/slug/:slug", h.GetEventBySlug)
		public.Get("/type/:type", h.GetEventsByType)
		public.Get("/", h.ListEvents)
		public.Get("/:id", h.GetEvent)
	}

	// ============================================================
	// 2. PROTECTED EVENT ROUTES
	// ============================================================
	protected := router.Group("/events")
	protected.Use(authMiddleware)
	protected.Use(authzMiddleware)
	{
		protected.Get("/", h.ListUserEvents)
		protected.Get("/me/search", h.SearchEvents)   // ⬅️ authenticated, permission-scoped search
		protected.Post("/", h.CreateEvent)
		protected.Post("/draft", h.CreateEventDraft)
		
		// AI-assisted draft generation
		protected.Post("/ai/generate-draft", h.GenerateEventDraft)

		bulk := protected.Group("/bulk")
		{
			bulk.Delete("/", h.BulkDeleteEvents)
			bulk.Delete("/permanent", h.BulkPermanentlyDeleteEvents)
			bulk.Post("/restore", h.BulkRestoreEvents)
			bulk.Post("/publish", h.BulkPublishEvents)
			bulk.Post("/cancel", h.BulkCancelEvents)
			bulk.Post("/complete", h.BulkCompleteEvents)
			bulk.Post("/duplicate", h.BulkDuplicateEvents)
			bulk.Delete("/media", h.BulkDeleteEventMedia)
		}

		protected.Put("/:id", h.UpdateEvent)
		protected.Delete("/:id", h.DeleteEvent)
		protected.Delete("/:id/permanent", h.PermanentlyDeleteEvent)
		protected.Post("/:id/restore", h.RestoreEvent)
		protected.Post("/:id/publish", h.PublishEvent)
		protected.Post("/:id/cancel", h.CancelEvent)
		protected.Post("/:id/complete", h.CompleteEvent)
		protected.Post("/:id/duplicate", h.DuplicateEvent)

		protected.Post("/:id/image", h.UploadEventImage)
		protected.Post("/:id/certificate", h.UploadCertificateTemplate)
		protected.Delete("/:id/image", h.DeleteEventImage)
		protected.Delete("/:id/certificate", h.DeleteEventCertificate)
		protected.Delete("/:id/media", h.DeleteAllEventMedia)
	}
}