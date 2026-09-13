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
	//
	// Specific literal paths first, then parameterized.
	// The /:id route is constrained to UUIDs so it doesn't swallow
	// names like /me that belong to the protected group.
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

		// Public feed
		public.Get("/", h.ListEvents)

		// Parameterized LAST, constrained to UUIDs.
		public.Get("/:id<guid>", h.GetEvent)
	}

	// ============================================================
	// 2. PROTECTED EVENT ROUTES
	// ============================================================
	//
	// /me is a literal path, so it's unambiguous with the UUID-only
	// /:id route above.
	protected := router.Group("/events")
	protected.Use(authMiddleware)
	protected.Use(authzMiddleware)
	{
		// ---- User-scoped listings ----
		protected.Get("/me", h.ListUserEvents)
		protected.Get("/me/search", h.SearchEvents)

		// ---- Create ----
		protected.Post("/", h.CreateEvent)
		protected.Post("/draft", h.CreateEventDraft)
		protected.Post("/ai/generate-draft", h.GenerateEventDraft)

		// ---- Bulk operations ----
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

		// ---- Single-event mutations (all UUID-constrained) ----
		protected.Put("/:id<guid>", h.UpdateEvent)
		protected.Delete("/:id<guid>", h.DeleteEvent)
		protected.Delete("/:id<guid>/permanent", h.PermanentlyDeleteEvent)
		protected.Post("/:id<guid>/restore", h.RestoreEvent)
		protected.Post("/:id<guid>/publish", h.PublishEvent)
		protected.Post("/:id<guid>/cancel", h.CancelEvent)
		protected.Post("/:id<guid>/complete", h.CompleteEvent)
		protected.Post("/:id<guid>/duplicate", h.DuplicateEvent)

		// ---- Media (UUID-constrained) ----
		protected.Post("/:id<guid>/image", h.UploadEventImage)
		protected.Post("/:id<guid>/certificate", h.UploadCertificateTemplate)
		protected.Delete("/:id<guid>/image", h.DeleteEventImage)
		protected.Delete("/:id<guid>/certificate", h.DeleteEventCertificate)
		protected.Delete("/:id<guid>/media", h.DeleteAllEventMedia)
	}
}