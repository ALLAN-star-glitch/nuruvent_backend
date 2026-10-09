package eventhandler

import (
	"github.com/gofiber/fiber/v3"
)

func (h *EventHandler) RegisterRoutes(
	router fiber.Router,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
	optionalAuthMiddleware fiber.Handler,
) {
	// ============================================================
	// IMPORTANT: static segments must be registered BEFORE any
	// dynamic /:id route, otherwise /events/me, /events/search,
	// /events/upcoming etc. will be swallowed by /events/:id.
	// ============================================================

	events := router.Group("/events")

	// ------------------------------------------------------------
	// PUBLIC STATIC ROUTES
	//
	// These endpoints are publicly reachable, but they behave
	// differently for authenticated callers (private events,
	// creator info, own-draft visibility). OptionalAuth populates
	// the context when a token is present, and silently falls
	// through to anonymous otherwise.
	// ------------------------------------------------------------
	events.Get("/upcoming", optionalAuthMiddleware, h.GetUpcomingEvents)
	events.Get("/past", optionalAuthMiddleware, h.GetPastEvents)
	events.Get("/search", optionalAuthMiddleware, h.SearchEvents)
	events.Get("/types", h.GetEventTypes)
	events.Get("/statuses", h.GetEventStatuses)
	events.Get("/categories", h.GetCategories)
	events.Get("/ticket-types", h.GetTicketTypes)
	events.Get("/slug/:slug", optionalAuthMiddleware, h.GetEventBySlug)
	events.Get("/type/:type", optionalAuthMiddleware, h.GetEventsByType)
	events.Get("/", optionalAuthMiddleware, h.ListEvents)

	
	// ------------------------------------------------------------
	// PROTECTED STATIC ROUTES (must come before /:id)
	// ------------------------------------------------------------
	events.Get("/me", authMiddleware, authzMiddleware, h.ListUserEvents)
	events.Get("/me/search", authMiddleware, authzMiddleware, h.SearchEvents)

	events.Post("/", authMiddleware, authzMiddleware, h.CreateEvent)
	events.Post("/draft", authMiddleware, authzMiddleware, h.CreateEventDraft)
	events.Post("/ai/generate-draft", authMiddleware, authzMiddleware, h.GenerateEventDraft)

	// ---- Bulk operations ----
	bulk := events.Group("/bulk")
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

	// ------------------------------------------------------------
	// SINGLE-EVENT MUTATIONS (dynamic /:id — protected only)
	// ------------------------------------------------------------
	events.Put("/:id", authMiddleware, authzMiddleware, h.UpdateEvent)
	events.Delete("/:id", authMiddleware, authzMiddleware, h.DeleteEvent)
	events.Delete("/:id/permanent", authMiddleware, authzMiddleware, h.PermanentlyDeleteEvent)
	events.Post("/:id/restore", authMiddleware, authzMiddleware, h.RestoreEvent)
	events.Post("/:id/publish", authMiddleware, authzMiddleware, h.PublishEvent)
	events.Post("/:id/cancel", authMiddleware, authzMiddleware, h.CancelEvent)
	events.Post("/:id/complete", authMiddleware, authzMiddleware, h.CompleteEvent)
	events.Post("/:id/duplicate", authMiddleware, authzMiddleware, h.DuplicateEvent)
	events.Post("/:id/schedules/reorder", authMiddleware, authzMiddleware, h.ReorderSchedules)

	// ---- Meeting management ----
	events.Post("/:id/meeting", authMiddleware, authzMiddleware, h.CreateMeeting)
	events.Delete("/:id/meeting", authMiddleware, authzMiddleware, h.DeleteMeeting)
	events.Post("/:id/meeting/regenerate", authMiddleware, authzMiddleware, h.RegenerateMeeting)

	// ---- Media ----
	events.Post("/:id/image", authMiddleware, authzMiddleware, h.UploadEventImage)
	events.Post("/:id/certificate", authMiddleware, authzMiddleware, h.UploadCertificateTemplate)
	events.Delete("/:id/image", authMiddleware, authzMiddleware, h.DeleteEventImage)
	events.Delete("/:id/certificate", authMiddleware, authzMiddleware, h.DeleteEventCertificate)
	events.Delete("/:id/media", authMiddleware, authzMiddleware, h.DeleteAllEventMedia)

	// ------------------------------------------------------------
	// PUBLIC DYNAMIC ROUTE — LAST
	//
	// GET /events/:id must accept both anonymous and authenticated
	// callers. Anonymous → public/unlisted events only.
	// Authenticated → also private events where the user is the
	// creator or a team member.
	// ------------------------------------------------------------
	events.Get("/:id", optionalAuthMiddleware, h.GetEvent)
}