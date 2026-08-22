// internal/modules/events/handler/routes.go

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
		// ✅ All public endpoints now include creator info
		public.Get("/upcoming", h.GetUpcomingEventsWithCreator)
		public.Get("/past", h.GetPastEvents)
		public.Get("/types", h.GetEventTypes)
		public.Get("/statuses", h.GetEventStatuses)
		public.Get("/search", h.SearchEvents)
		public.Get("/", h.ListEvents)
		public.Get("/slug/:slug", h.GetEventBySlugWithCreator)
		public.Get("/type/:type", h.GetEventsByType)
		public.Get("/:id", h.GetEventByIDWithCreator)
	}

	// ============================================================
	// PROTECTED ROUTES - Account Level (Auth required)
	// ============================================================
	protected := router.Group("/accounts/:accountId/events")
	protected.Use(authMiddleware)
	protected.Use(authzMiddleware)
	{
		// ✅ All account endpoints now include creator info
		protected.Get("/", h.GetEventsByAccountWithCreator)
		protected.Post("/draft", h.CreateDraft)
		protected.Post("/", h.CreateEvent)
		protected.Post("/:eventId/image", h.UploadEventImage)
		protected.Post("/:eventId/certificate", h.UploadCertificateTemplate)
		protected.Delete("/:eventId/image", h.DeleteEventImage)
		protected.Delete("/:eventId/certificate", h.DeleteEventCertificate)
		protected.Delete("/:eventId/media", h.DeleteAllEventMedia)
		protected.Delete("/bulk/media", h.BulkDeleteEventMedia)
	}

	// ============================================================
	// ✅ BULK OPERATIONS - REGISTER FIRST (before single routes)
	// ============================================================
	bulkRoutes := router.Group("/events/bulk")
	bulkRoutes.Use(authMiddleware)
	bulkRoutes.Use(authzMiddleware)
	{
		bulkRoutes.Delete("/", h.BulkDeleteEvents)
		bulkRoutes.Delete("/permanent", h.BulkPermanentlyDeleteEvents)
		bulkRoutes.Post("/restore", h.BulkRestoreEvents)
		bulkRoutes.Post("/publish", h.BulkPublishEvents)
		bulkRoutes.Post("/cancel", h.BulkCancelEvents)
		bulkRoutes.Post("/complete", h.BulkCompleteEvents)
		bulkRoutes.Post("/duplicate", h.BulkDuplicateEvents)
	}

	// ============================================================
	// ✅ SINGLE EVENT MANAGEMENT - REGISTER AFTER bulk routes
	// ============================================================
	eventRoutes := router.Group("/events")
	eventRoutes.Use(authMiddleware)
	eventRoutes.Use(authzMiddleware)
	{
		eventRoutes.Put("/:id", h.UpdateEvent)
		eventRoutes.Delete("/:id", h.DeleteEvent)
		eventRoutes.Delete("/:id/permanent", h.PermanentlyDeleteEvent)
		eventRoutes.Post("/:id/restore", h.RestoreEvent)
		eventRoutes.Post("/:id/publish", h.PublishEvent)
		eventRoutes.Post("/:id/cancel", h.CancelEvent)
		eventRoutes.Post("/:id/complete", h.CompleteEvent)
		eventRoutes.Post("/:id/duplicate", h.DuplicateEvent)
	}
}