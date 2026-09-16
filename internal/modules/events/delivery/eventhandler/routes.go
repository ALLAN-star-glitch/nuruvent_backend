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
	// PUBLIC ROUTES
	// ============================================================
	public := router.Group("/events")
	{
		public.Get("/upcoming", h.GetUpcomingEvents)
		public.Get("/past", h.GetPastEvents)
		public.Get("/search", h.SearchEvents)
		public.Get("/types", h.GetEventTypes)
		public.Get("/statuses", h.GetEventStatuses)
		public.Get("/categories", h.GetCategories)
		public.Get("/ticket-types", h.GetTicketTypes)
		public.Get("/slug/:slug", h.GetEventBySlug)
		public.Get("/type/:type", h.GetEventsByType)
		public.Get("/", h.ListEvents)
		public.Get("/:id<guid>", h.GetEvent)
	}

	// ============================================================
	// PROTECTED EVENT ROUTES
	// ============================================================
	//
	// Auth is applied PER-ROUTE. Do NOT use protected.Use(...) here —
	// group-level middleware applies by path prefix and would intercept
	// /events/:id/register and other cross-module routes.
	protected := router.Group("/events")
	{
		// ---- User-scoped listings ----
		protected.Get("/me", authMiddleware, authzMiddleware, h.ListUserEvents)
		protected.Get("/me/search", authMiddleware, authzMiddleware, h.SearchEvents)

		// ---- Create ----
		protected.Post("/", authMiddleware, authzMiddleware, h.CreateEvent)
		protected.Post("/draft", authMiddleware, authzMiddleware, h.CreateEventDraft)
		protected.Post("/ai/generate-draft", authMiddleware, authzMiddleware, h.GenerateEventDraft)

		// ---- Bulk operations ----
		bulk := protected.Group("/bulk")
		{
			bulk.Delete("/", authMiddleware, authzMiddleware, h.BulkDeleteEvents)
			bulk.Delete("/permanent", authMiddleware, authzMiddleware, h.BulkPermanentlyDeleteEvents)
			bulk.Post("/restore", authMiddleware, authzMiddleware, h.BulkRestoreEvents)
			bulk.Post("/publish", authMiddleware, authzMiddleware, h.BulkPublishEvents)
			bulk.Post("/cancel", authMiddleware, authzMiddleware, h.BulkCancelEvents)
			bulk.Post("/complete", authMiddleware, authzMiddleware, h.BulkCompleteEvents)
			bulk.Post("/duplicate", authMiddleware, authzMiddleware, h.BulkDuplicateEvents)
			bulk.Delete("/media", authMiddleware, authzMiddleware, h.BulkDeleteEventMedia)
		}

		// ---- Single-event mutations ----
		protected.Put("/:id<guid>", authMiddleware, authzMiddleware, h.UpdateEvent)
		protected.Delete("/:id<guid>", authMiddleware, authzMiddleware, h.DeleteEvent)
		protected.Delete("/:id<guid>/permanent", authMiddleware, authzMiddleware, h.PermanentlyDeleteEvent)
		protected.Post("/:id<guid>/restore", authMiddleware, authzMiddleware, h.RestoreEvent)
		protected.Post("/:id<guid>/publish", authMiddleware, authzMiddleware, h.PublishEvent)
		protected.Post("/:id<guid>/cancel", authMiddleware, authzMiddleware, h.CancelEvent)
		protected.Post("/:id<guid>/complete", authMiddleware, authzMiddleware, h.CompleteEvent)
		protected.Post("/:id<guid>/duplicate", authMiddleware, authzMiddleware, h.DuplicateEvent)

		// ---- Media ----
		protected.Post("/:id<guid>/image", authMiddleware, authzMiddleware, h.UploadEventImage)
		protected.Post("/:id<guid>/certificate", authMiddleware, authzMiddleware, h.UploadCertificateTemplate)
		protected.Delete("/:id<guid>/image", authMiddleware, authzMiddleware, h.DeleteEventImage)
		protected.Delete("/:id<guid>/certificate", authMiddleware, authzMiddleware, h.DeleteEventCertificate)
		protected.Delete("/:id<guid>/media", authMiddleware, authzMiddleware, h.DeleteAllEventMedia)
	}
}