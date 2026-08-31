// internal/modules/events/handler/routes.go

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
	// PUBLIC ROUTES (No auth required)
	// ============================================================
	public := router.Group("/events")
	{
		// ✅ MOST SPECIFIC FIRST: Search
		public.Get("/search", h.SearchEvents)
		
		// ✅ Static routes (no parameters)
		public.Get("/upcoming", h.GetUpcomingEvents)
		public.Get("/past", h.GetPastEvents)
		public.Get("/types", h.GetEventTypes)
		public.Get("/statuses", h.GetEventStatuses)
		public.Get("/", h.ListEvents)
		
		// ✅ With creator info - require auth
		public.Get("/upcoming/with-creator", authMiddleware, h.GetUpcomingEventsWithCreator)
		
		// ✅ Routes with specific parameter patterns
		public.Get("/slug/:slug", h.GetEventBySlug)
		public.Get("/slug/:slug/with-creator", authMiddleware, h.GetEventBySlugWithCreator)
		
		// ✅ Type routes
		public.Get("/type/:type", h.GetEventsByType)
		
		// ✅ LEAST SPECIFIC: ID routes (MUST BE LAST!)
		public.Get("/:id", h.GetEvent)
		public.Get("/:id/with-creator", authMiddleware, h.GetEventByIDWithCreator)
	}

	// ============================================================
	// PERSONAL ROUTES - User's own personal team events
	// ============================================================
	personal := router.Group("/users/me/events")
	personal.Use(authMiddleware)
	personal.Use(authzMiddleware)
	{
		// READ
		personal.Get("/", h.GetMyEvents)
		personal.Get("/with-creator", h.GetMyEventsWithCreator)
		
		// CREATE
		personal.Post("/draft", h.CreatePersonalDraft)
		personal.Post("/", h.CreatePersonalEvent)
		
		// ✅ BULK OPERATIONS (MUST come before /:id)
		personal.Delete("/bulk", h.BulkDeleteEvents)
		personal.Delete("/bulk/permanent", h.BulkPermanentlyDeleteEvents)
		personal.Post("/bulk/restore", h.BulkRestoreEvents)
		personal.Post("/bulk/publish", h.BulkPublishEvents)
		personal.Post("/bulk/cancel", h.BulkCancelEvents)
		personal.Post("/bulk/complete", h.BulkCompleteEvents)
		personal.Post("/bulk/duplicate", h.BulkDuplicateEvents)
		personal.Delete("/bulk/media", h.BulkDeleteEventMedia)
		
		// ✅ MEDIA (MUST come before /:id)
		personal.Post("/:eventId/image", h.UploadEventImage)
		personal.Post("/:eventId/certificate", h.UploadCertificateTemplate)
		personal.Delete("/:eventId/image", h.DeleteEventImage)
		personal.Delete("/:eventId/certificate", h.DeleteEventCertificate)
		personal.Delete("/:eventId/media", h.DeleteAllEventMedia)
		
		// ✅ SINGLE EVENT OPERATIONS (MUST be LAST)
		personal.Put("/:id", h.UpdateEvent)
		personal.Delete("/:id", h.DeleteEvent)
		personal.Delete("/:id/permanent", h.PermanentlyDeleteEvent)
		personal.Post("/:id/restore", h.RestoreEvent)
		personal.Post("/:id/publish", h.PublishEvent)
		personal.Post("/:id/cancel", h.CancelEvent)
		personal.Post("/:id/complete", h.CompleteEvent)
		personal.Post("/:id/duplicate", h.DuplicateEvent)
	}

	// ============================================================
	// USER ROUTES - View events by specific user's personal team
	// ============================================================
	userRoutes := router.Group("/users/:userId/events")
	userRoutes.Use(authMiddleware)
	userRoutes.Use(authzMiddleware)
	{
		userRoutes.Get("/with-creator", h.GetEventsByUserWithCreator)
	}

	// ============================================================
	// INSTITUTION ROUTES - Public viewing (no auth required for basic info)
	// ============================================================
	institutionRoutes := router.Group("/institutions/:institutionId/events")
	{
		institutionRoutes.Get("/", h.GetEventsByInstitution)
		institutionRoutes.Get("/with-creator", authMiddleware, h.GetEventsByInstitutionWithCreator)
	}

	// ============================================================
	// INSTITUTION MANAGEMENT - Protected (Auth + Authz required)
	// FULL CRUD - Same as personal team!
	// ============================================================
	institutionManagement := router.Group("/institutions/:institutionId/events")
	institutionManagement.Use(authMiddleware)
	institutionManagement.Use(authzMiddleware)
	{
		// READ
		institutionManagement.Get("/", h.GetEventsByInstitution)
		institutionManagement.Get("/with-creator", h.GetEventsByInstitutionWithCreator)
		
		// CREATE
		institutionManagement.Post("/draft", h.CreateDraft)
		institutionManagement.Post("/", h.CreateEvent)
		
		// ✅ BULK OPERATIONS (MUST come before /:id)
		institutionManagement.Delete("/bulk", h.BulkDeleteEvents)
		institutionManagement.Delete("/bulk/permanent", h.BulkPermanentlyDeleteEvents)
		institutionManagement.Post("/bulk/restore", h.BulkRestoreEvents)
		institutionManagement.Post("/bulk/publish", h.BulkPublishEvents)
		institutionManagement.Post("/bulk/cancel", h.BulkCancelEvents)
		institutionManagement.Post("/bulk/complete", h.BulkCompleteEvents)
		institutionManagement.Post("/bulk/duplicate", h.BulkDuplicateEvents)
		institutionManagement.Delete("/bulk/media", h.BulkDeleteEventMedia)
		
		// ✅ MEDIA (MUST come before /:id)
		institutionManagement.Post("/:eventId/image", h.UploadEventImage)
		institutionManagement.Post("/:eventId/certificate", h.UploadCertificateTemplate)
		institutionManagement.Delete("/:eventId/image", h.DeleteEventImage)
		institutionManagement.Delete("/:eventId/certificate", h.DeleteEventCertificate)
		institutionManagement.Delete("/:eventId/media", h.DeleteAllEventMedia)
		
		// ✅ SINGLE EVENT OPERATIONS (MUST be LAST)
		institutionManagement.Put("/:id", h.UpdateEvent)
		institutionManagement.Delete("/:id", h.DeleteEvent)
		institutionManagement.Delete("/:id/permanent", h.PermanentlyDeleteEvent)
		institutionManagement.Post("/:id/restore", h.RestoreEvent)
		institutionManagement.Post("/:id/publish", h.PublishEvent)
		institutionManagement.Post("/:id/cancel", h.CancelEvent)
		institutionManagement.Post("/:id/complete", h.CompleteEvent)
		institutionManagement.Post("/:id/duplicate", h.DuplicateEvent)
	}

	// ============================================================
	// BULK OPERATIONS - For any team (Auth + Authz required)
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
	// SINGLE EVENT MANAGEMENT - For any team (Auth + Authz required)
	// ============================================================
	eventRoutes := router.Group("/events")
	eventRoutes.Use(authMiddleware)
	eventRoutes.Use(authzMiddleware)
	{
		// ✅ BULK OPERATIONS (MUST come before /:id)
		eventRoutes.Delete("/bulk", h.BulkDeleteEvents)
		eventRoutes.Delete("/bulk/permanent", h.BulkPermanentlyDeleteEvents)
		eventRoutes.Post("/bulk/restore", h.BulkRestoreEvents)
		eventRoutes.Post("/bulk/publish", h.BulkPublishEvents)
		eventRoutes.Post("/bulk/cancel", h.BulkCancelEvents)
		eventRoutes.Post("/bulk/complete", h.BulkCompleteEvents)
		eventRoutes.Post("/bulk/duplicate", h.BulkDuplicateEvents)
		
		// ✅ SINGLE EVENT OPERATIONS (MUST be LAST)
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