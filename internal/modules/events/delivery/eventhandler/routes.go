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
		// Simple versions (no creator info)
		public.Get("/upcoming", h.GetUpcomingEvents)
		public.Get("/slug/:slug", h.GetEventBySlug)
		public.Get("/:id", h.GetEvent)

		// With creator info versions - require auth (auth middleware only)
		public.Get("/upcoming/with-creator", authMiddleware, h.GetUpcomingEventsWithCreator)
		public.Get("/slug/:slug/with-creator", authMiddleware, h.GetEventBySlugWithCreator)
		public.Get("/:id/with-creator", authMiddleware, h.GetEventByIDWithCreator)

		// Other public endpoints
		public.Get("/past", h.GetPastEvents)
		public.Get("/types", h.GetEventTypes)
		public.Get("/statuses", h.GetEventStatuses)
		public.Get("/search", h.SearchEvents)
		public.Get("/", h.ListEvents)
		public.Get("/type/:type", h.GetEventsByType)
	}

	// ============================================================
	// PERSONAL ROUTES - User's own personal team events (Auth + Authz required)
	// ONLY for events owned by the user's personal team
	// ============================================================
	personal := router.Group("/users/me/events")
	personal.Use(authMiddleware)
	personal.Use(authzMiddleware)
	{
		// READ - Get events from user's personal team
		personal.Get("/", h.GetMyEvents)
		personal.Get("/with-creator", h.GetMyEventsWithCreator)
		
		// CREATE - Create events in user's personal team
		personal.Post("/draft", h.CreatePersonalDraft)
		personal.Post("/", h.CreatePersonalEvent)
		
		// UPDATE
		personal.Put("/:id", h.UpdateEvent)
		
		// DELETE
		personal.Delete("/:id", h.DeleteEvent)
		personal.Delete("/:id/permanent", h.PermanentlyDeleteEvent)
		personal.Post("/:id/restore", h.RestoreEvent)
		
		// STATUS
		personal.Post("/:id/publish", h.PublishEvent)
		personal.Post("/:id/cancel", h.CancelEvent)
		personal.Post("/:id/complete", h.CompleteEvent)
		
		// DUPLICATE
		personal.Post("/:id/duplicate", h.DuplicateEvent)
		
		// MEDIA - Upload (certificates included)
		personal.Post("/:eventId/image", h.UploadEventImage)
		personal.Post("/:eventId/certificate", h.UploadCertificateTemplate)
		
		// MEDIA - Delete
		personal.Delete("/:eventId/image", h.DeleteEventImage)
		personal.Delete("/:eventId/certificate", h.DeleteEventCertificate)
		personal.Delete("/:eventId/media", h.DeleteAllEventMedia)
		
		// BULK - For personal team
		personal.Delete("/bulk", h.BulkDeleteEvents)
		personal.Delete("/bulk/permanent", h.BulkPermanentlyDeleteEvents)
		personal.Post("/bulk/restore", h.BulkRestoreEvents)
		personal.Post("/bulk/publish", h.BulkPublishEvents)
		personal.Post("/bulk/cancel", h.BulkCancelEvents)
		personal.Post("/bulk/complete", h.BulkCompleteEvents)
		personal.Post("/bulk/duplicate", h.BulkDuplicateEvents)
		personal.Delete("/bulk/media", h.BulkDeleteEventMedia)
	}

	// ============================================================
	// USER ROUTES - View events by specific user's personal team
	// (Auth + Authz required to see creator info)
	// ============================================================
	userRoutes := router.Group("/users/:userId/events")
	userRoutes.Use(authMiddleware)
	userRoutes.Use(authzMiddleware)
	{
		userRoutes.Get("/with-creator", h.GetEventsByUserWithCreator)
	}

	// ============================================================
	// INSTITUTION ROUTES - For institution members
	// ============================================================
	// Public viewing (no auth required for basic info)
	institutionRoutes := router.Group("/institutions/:institutionId/events")
	{
		// GET endpoints - Public, but filters private events
		institutionRoutes.Get("/", h.GetEventsByInstitution)
		
		// With creator info - requires auth to see creator details
		institutionRoutes.Get("/with-creator", authMiddleware, h.GetEventsByInstitutionWithCreator)
	}

	// ============================================================
	// INSTITUTION MANAGEMENT - Protected (Auth + Authz required)
	// Full CRUD operations for institution events
	// ============================================================
	institutionManagement := router.Group("/institutions/:institutionId/events")
	institutionManagement.Use(authMiddleware)
	institutionManagement.Use(authzMiddleware)
	{
		// CREATE
		institutionManagement.Post("/draft", h.CreateDraft)
		institutionManagement.Post("/", h.CreateEvent)
		
		// UPDATE
		institutionManagement.Put("/:id", h.UpdateEvent)
		
		// DELETE
		institutionManagement.Delete("/:id", h.DeleteEvent)
		institutionManagement.Delete("/:id/permanent", h.PermanentlyDeleteEvent)
		institutionManagement.Post("/:id/restore", h.RestoreEvent)
		
		// STATUS
		institutionManagement.Post("/:id/publish", h.PublishEvent)
		institutionManagement.Post("/:id/cancel", h.CancelEvent)
		institutionManagement.Post("/:id/complete", h.CompleteEvent)
		
		// DUPLICATE
		institutionManagement.Post("/:id/duplicate", h.DuplicateEvent)
		
		// MEDIA - Upload (certificates included)
		institutionManagement.Post("/:eventId/image", h.UploadEventImage)
		institutionManagement.Post("/:eventId/certificate", h.UploadCertificateTemplate)
		
		// MEDIA - Delete
		institutionManagement.Delete("/:eventId/image", h.DeleteEventImage)
		institutionManagement.Delete("/:eventId/certificate", h.DeleteEventCertificate)
		institutionManagement.Delete("/:eventId/media", h.DeleteAllEventMedia)
		
		// BULK - For institution team
		institutionManagement.Delete("/bulk", h.BulkDeleteEvents)
		institutionManagement.Delete("/bulk/permanent", h.BulkPermanentlyDeleteEvents)
		institutionManagement.Post("/bulk/restore", h.BulkRestoreEvents)
		institutionManagement.Post("/bulk/publish", h.BulkPublishEvents)
		institutionManagement.Post("/bulk/cancel", h.BulkCancelEvents)
		institutionManagement.Post("/bulk/complete", h.BulkCompleteEvents)
		institutionManagement.Post("/bulk/duplicate", h.BulkDuplicateEvents)
		institutionManagement.Delete("/bulk/media", h.BulkDeleteEventMedia)
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