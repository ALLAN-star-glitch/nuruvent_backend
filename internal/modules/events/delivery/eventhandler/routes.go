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

		// With creator info versions
		public.Get("/upcoming/with-creator", h.GetUpcomingEventsWithCreator)
		public.Get("/slug/:slug/with-creator", h.GetEventBySlugWithCreator)
		public.Get("/:id/with-creator", h.GetEventByIDWithCreator)

		// Other public endpoints
		public.Get("/past", h.GetPastEvents)
		public.Get("/types", h.GetEventTypes)
		public.Get("/statuses", h.GetEventStatuses)
		public.Get("/search", h.SearchEvents)
		public.Get("/", h.ListEvents)
		public.Get("/type/:type", h.GetEventsByType)
	}

	// ============================================================
	// PERSONAL ROUTES - User's own events (Auth required)
	// ============================================================
	personal := router.Group("/users/me/events")
	personal.Use(authMiddleware)
	{
		personal.Get("/", h.GetMyEvents)
		personal.Get("/with-creator", h.GetMyEventsWithCreator)
		personal.Post("/draft", h.CreatePersonalDraft)
		personal.Post("/", h.CreatePersonalEvent)
	}

	// ============================================================
	// USER ROUTES - View events by specific user (Auth + Authz required)
	// ============================================================
	userRoutes := router.Group("/users/:userId/events")
	userRoutes.Use(authMiddleware)
	userRoutes.Use(authzMiddleware)
	{
		userRoutes.Get("/with-creator", h.GetEventsByUserWithCreator)
	}

	// ============================================================
	// INSTITUTION ROUTES - Public viewing (Auth optional, authz at service level)
	// ============================================================
	institutionRoutes := router.Group("/institutions/:institutionId/events")
	{
		// GET endpoints - Public, but filters private events
		institutionRoutes.Get("/", h.GetEventsByInstitution)
		institutionRoutes.Get("/with-creator", h.GetEventsByInstitutionWithCreator)
	}

	// ============================================================
	// INSTITUTION MANAGEMENT - Protected (Auth + Authz required)
	// ============================================================
	institutionManagement := router.Group("/institutions/:institutionId/events")
	institutionManagement.Use(authMiddleware)
	institutionManagement.Use(authzMiddleware)
	{
		// POST endpoints - require write permissions
		institutionManagement.Post("/draft", h.CreateDraft)
		institutionManagement.Post("/", h.CreateEvent)
		
		// Media uploads - require write permissions
		institutionManagement.Post("/:eventId/image", h.UploadEventImage)
		institutionManagement.Post("/:eventId/certificate", h.UploadCertificateTemplate)
		
		// Media deletions - require write permissions
		institutionManagement.Delete("/:eventId/image", h.DeleteEventImage)
		institutionManagement.Delete("/:eventId/certificate", h.DeleteEventCertificate)
		institutionManagement.Delete("/:eventId/media", h.DeleteAllEventMedia)
		institutionManagement.Delete("/bulk/media", h.BulkDeleteEventMedia)
	}

	// ============================================================
	// BULK OPERATIONS - REGISTER FIRST (before single routes)
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
	// SINGLE EVENT MANAGEMENT - REGISTER AFTER bulk routes
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