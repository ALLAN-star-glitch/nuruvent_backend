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
		public.Get("/upcoming", h.GetUpcomingEvents)
		public.Get("/past", h.GetPastEvents)
		public.Get("/types", h.GetEventTypes)
		public.Get("/statuses", h.GetEventStatuses)
		public.Get("/search", h.SearchEvents)
		public.Get("/", h.ListEvents)
		// ⚠️ Move these specific routes before the wildcard
		public.Get("/slug/:slug", h.GetEventBySlug)
		public.Get("/type/:type", h.GetEventsByType)
		// ⚠️ Wildcard route should be LAST
		public.Get("/:id", h.GetEvent)
	}

	// ============================================================
	// PROTECTED ROUTES - Account Level (Auth required)
	// ============================================================
	protected := router.Group("/accounts/:accountId/events")
	protected.Use(authMiddleware)
	protected.Use(authzMiddleware)
	{
		protected.Get("/", h.GetEventsByAccount)
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
		bulkRoutes.Delete("/", h.BulkDeleteEvents)              // DELETE /events/bulk/
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
		eventRoutes.Delete("/:id", h.DeleteEvent)               // DELETE /events/{id}
		eventRoutes.Delete("/:id/permanent", h.PermanentlyDeleteEvent)
		eventRoutes.Post("/:id/restore", h.RestoreEvent)
		eventRoutes.Post("/:id/publish", h.PublishEvent)
		eventRoutes.Post("/:id/cancel", h.CancelEvent)
		eventRoutes.Post("/:id/complete", h.CompleteEvent)
		eventRoutes.Post("/:id/duplicate", h.DuplicateEvent)
	}
}