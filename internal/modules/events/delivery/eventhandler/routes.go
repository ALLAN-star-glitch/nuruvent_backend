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
	// 1. PUBLIC ROUTES - Accessible without authentication
	//    Purpose: Public event discovery, SEO, sharing, and metadata
	// ============================================================
	public := router.Group("/events")
	{
		// Unified event listing with filters (team_id, team_type, status, etc.)
		// Used by: Team switcher, advanced search, public listings
		public.Get("/", h.ListEvents)

		// ---- SEARCH (Requires Auth + Authz) ----
		// Authenticated users can search public + private events they have access to
		public.Get("/search", authMiddleware, authzMiddleware, h.SearchEvents)

		// Convenience endpoints for common queries
		public.Get("/upcoming", h.GetUpcomingEvents) // Events happening soon
		public.Get("/past", h.GetPastEvents)         // Events that have ended

		// Reference data (no team/scope needed)
		public.Get("/types", h.GetEventTypes)       // All event types (workshop, webinar, etc.)
		public.Get("/statuses", h.GetEventStatuses) // All statuses (draft, published, etc.)
		public.Get("/categories", h.GetCategories)

		// Single event access (public visibility)
		public.Get("/slug/:slug", h.GetEventBySlug)     // SEO-friendly URLs
		public.Get("/type/:type", h.GetEventsByType)    // All events of a specific type

		// MUST BE LAST - catches any /events/{id} requests
		public.Get("/:id", h.GetEvent) // Get event by UUID
	}

	// ============================================================
	// 2. PERSONAL TEAM ROUTES - User's own team
	//    Purpose: Full CRUD operations for user's personal team
	//    Scope: personal:team:{user_id}
	// ============================================================
	personal := router.Group("/users/me/events")
	personal.Use(authMiddleware)  // User must be logged in
	personal.Use(authzMiddleware) // User must have permission
	{
		// ---- READ ----
		personal.Get("/", h.GetMyEvents) // List personal team events

		// ---- CREATE ----
		personal.Post("/", h.CreatePersonalEvent)     // Create published event
		personal.Post("/draft", h.CreatePersonalDraft) // Create draft event

		// ---- BULK OPERATIONS ----
		personal.Delete("/bulk", h.BulkDeleteEvents)                  // Soft delete multiple events
		personal.Delete("/bulk/permanent", h.BulkPermanentlyDeleteEvents) // Hard delete
		personal.Post("/bulk/restore", h.BulkRestoreEvents)           // Restore soft-deleted
		personal.Post("/bulk/publish", h.BulkPublishEvents)           // Publish multiple
		personal.Post("/bulk/cancel", h.BulkCancelEvents)             // Cancel multiple
		personal.Post("/bulk/complete", h.BulkCompleteEvents)         // Complete multiple
		personal.Post("/bulk/duplicate", h.BulkDuplicateEvents)       // Duplicate multiple
		personal.Delete("/bulk/media", h.BulkDeleteEventMedia)        // Delete media for multiple

		// ---- MEDIA OPERATIONS ----
		personal.Post("/:eventId/image", h.UploadEventImage)               // Upload event image
		personal.Post("/:eventId/certificate", h.UploadCertificateTemplate) // Upload cert template
		personal.Delete("/:eventId/image", h.DeleteEventImage)             // Delete event image
		personal.Delete("/:eventId/certificate", h.DeleteEventCertificate) // Delete cert
		personal.Delete("/:eventId/media", h.DeleteAllEventMedia)          // Delete all media

		// ---- SINGLE EVENT OPERATIONS ----
		personal.Put("/:id", h.UpdateEvent)                   // Update event
		personal.Delete("/:id", h.DeleteEvent)                // Soft delete
		personal.Delete("/:id/permanent", h.PermanentlyDeleteEvent) // Hard delete
		personal.Post("/:id/restore", h.RestoreEvent)         // Restore
		personal.Post("/:id/publish", h.PublishEvent)         // Publish
		personal.Post("/:id/cancel", h.CancelEvent)           // Cancel
		personal.Post("/:id/complete", h.CompleteEvent)       // Complete
		personal.Post("/:id/duplicate", h.DuplicateEvent)     // Duplicate
	}

	// ============================================================
	// 3. TEAM ROUTES - Team-specific event operations
	//    Purpose: Full CRUD operations for any team
	//    Scope: team:{team_id}
	//    Same operations as personal team, but for any team
	// ============================================================
	team := router.Group("/teams/:teamId/events")
	team.Use(authMiddleware)
	team.Use(authzMiddleware)
	{
		// ---- READ ----
		team.Get("/", h.GetEventsByTeam) // List team events

		// ---- CREATE ----
		team.Post("/", h.CreateTeamEvent)     // Create published event
		team.Post("/draft", h.CreateTeamDraft) // Create draft event

		// ---- BULK OPERATIONS ----
		team.Delete("/bulk", h.BulkDeleteEvents)
		team.Delete("/bulk/permanent", h.BulkPermanentlyDeleteEvents)
		team.Post("/bulk/restore", h.BulkRestoreEvents)
		team.Post("/bulk/publish", h.BulkPublishEvents)
		team.Post("/bulk/cancel", h.BulkCancelEvents)
		team.Post("/bulk/complete", h.BulkCompleteEvents)
		team.Post("/bulk/duplicate", h.BulkDuplicateEvents)
		team.Delete("/bulk/media", h.BulkDeleteEventMedia)

		// ---- MEDIA OPERATIONS ----
		team.Post("/:eventId/image", h.UploadEventImage)
		team.Post("/:eventId/certificate", h.UploadCertificateTemplate)
		team.Delete("/:eventId/image", h.DeleteEventImage)
		team.Delete("/:eventId/certificate", h.DeleteEventCertificate)
		team.Delete("/:eventId/media", h.DeleteAllEventMedia)

		// ---- SINGLE EVENT OPERATIONS ----
		team.Put("/:id", h.UpdateEvent)
		team.Delete("/:id", h.DeleteEvent)
		team.Delete("/:id/permanent", h.PermanentlyDeleteEvent)
		team.Post("/:id/restore", h.RestoreEvent)
		team.Post("/:id/publish", h.PublishEvent)
		team.Post("/:id/cancel", h.CancelEvent)
		team.Post("/:id/complete", h.CompleteEvent)
		team.Post("/:id/duplicate", h.DuplicateEvent)
	}

	// ============================================================
	// 4. BULK OPERATIONS - Cross-team bulk actions
	//    Purpose: Perform bulk operations on events across any team
	//    Used when user has permission across multiple teams
	// ============================================================
	bulk := router.Group("/events/bulk")
	bulk.Use(authMiddleware)
	bulk.Use(authzMiddleware)
	{
		bulk.Delete("/", h.BulkDeleteEvents)                  // Soft delete
		bulk.Delete("/permanent", h.BulkPermanentlyDeleteEvents) // Hard delete
		bulk.Post("/restore", h.BulkRestoreEvents)           // Restore
		bulk.Post("/publish", h.BulkPublishEvents)           // Publish
		bulk.Post("/cancel", h.BulkCancelEvents)             // Cancel
		bulk.Post("/complete", h.BulkCompleteEvents)         // Complete
		bulk.Post("/duplicate", h.BulkDuplicateEvents)       // Duplicate
		bulk.Delete("/media", h.BulkDeleteEventMedia)        // Delete media
	}

	// ============================================================
	// 5. SINGLE EVENT OPERATIONS - Cross-team single event actions
	//    Purpose: Perform operations on a single event regardless of team
	//    Used when user has permission to manage the specific event
	// ============================================================
	single := router.Group("/events")
	single.Use(authMiddleware)
	single.Use(authzMiddleware)
	{
		single.Put("/:id", h.UpdateEvent)                   // Update
		single.Delete("/:id", h.DeleteEvent)                // Soft delete
		single.Delete("/:id/permanent", h.PermanentlyDeleteEvent) // Hard delete
		single.Post("/:id/restore", h.RestoreEvent)         // Restore
		single.Post("/:id/publish", h.PublishEvent)         // Publish
		single.Post("/:id/cancel", h.CancelEvent)           // Cancel
		single.Post("/:id/complete", h.CompleteEvent)       // Complete
		single.Post("/:id/duplicate", h.DuplicateEvent)     // Duplicate
	}
}